package workers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"time"

	ingPorts "riksdagskollen/internal/ingestion/ports"
	"riksdagskollen/internal/motions"
	"riksdagskollen/internal/motions/domain"
	motionPorts "riksdagskollen/internal/motions/ports"
)

// OldestBackfillRiksmote is where the history of voteringar begins.
const OldestBackfillRiksmote = "2002/03"

// betankandeBatch is how many pending betänkanden are asked for at a time.
const betankandeBatch = 25

// BackfillOptions steers one run of the operator backfill.
type BackfillOptions struct {
	// Riksmoten limits the run to these. Empty means every riksmöte with
	// voteringar (or declared in a mandate period), newest first, back to From.
	Riksmoten []string
	// From is the oldest riksmöte to take; default 2002/03.
	From string
	// Force redoes riksmöte × source rows already recorded as complete.
	Force bool
	// Sources limits the run to some of domain.BackfillMotions and
	// domain.BackfillBetankanden. Empty means both.
	Sources []string
	// SkipRelink leaves out the recompute of the yrkande links (and with them
	// every votering's origin) after the walk.
	SkipRelink bool
}

// MotionsBackfill is the operator's one-riksmöte-at-a-time ingestion of
// motions and betänkanden. It is never registered with the scheduler.
//
// Resumable: progress is recorded per riksmöte and source, and a restart skips
// what is complete. Idempotent: every write is the daily worker's own upsert.
// Polite: it pauses FetchDelay between requests, as the workers do.
type MotionsBackfill struct {
	svc     *motions.Service
	periods ingPorts.MandatePeriodReader
	now     func() time.Time
}

func NewMotionsBackfill(svc *motions.Service, periods ingPorts.MandatePeriodReader) *MotionsBackfill {
	return &MotionsBackfill{svc: svc, periods: periods}
}

// SetNow pins the clock that decides which riksmöten have opened.
func (b *MotionsBackfill) SetNow(now func() time.Time) { b.now = now }

func (b *MotionsBackfill) at() time.Time {
	if b.now != nil {
		return b.now().UTC()
	}
	return time.Now().UTC()
}

// BackfillRow is one riksmöte × source of a run.
type BackfillRow struct {
	Riksmote string
	Source   string
	Expected int
	Held     int
	Complete bool
	Skipped  bool // already complete from an earlier run
	Err      string
}

// Report is what a run, or a report-only invocation, found.
type Report struct {
	Rows        []BackfillRow
	Vocabulary  []VocabularyRow
	Voteringar  int
	WithOrigin  int
	Relinked    int
	Relink      bool
	Failed      int
	Interrupted bool
}

// VocabularyRow is one decision value seen, with how the site shows it.
type VocabularyRow struct {
	Field  string
	Value  string
	Count  int
	Shown  string
	Mapped bool
}

// Run walks the riksmöten newest first. It returns the report even when the
// context is cancelled part-way; Interrupted says so.
func (b *MotionsBackfill) Run(ctx context.Context, opts BackfillOptions) (*Report, error) {
	rms, err := b.scope(ctx, opts)
	if err != nil {
		return nil, err
	}
	progress, err := b.svc.BackfillProgress(ctx)
	if err != nil {
		return nil, err
	}
	done := map[string]bool{}
	for _, p := range progress {
		if p.CompletedAt != nil {
			done[p.Riksmote+"|"+p.Source] = true
		}
	}

	sources := opts.Sources
	if len(sources) == 0 {
		sources = []string{domain.BackfillMotions, domain.BackfillBetankanden}
	}
	rep := &Report{Relink: !opts.SkipRelink}
	mw := &MotionsWorker{svc: b.svc}
	for _, rm := range rms {
		if ctx.Err() != nil {
			rep.Interrupted = true
			break
		}
		for _, src := range sources {
			if ctx.Err() != nil {
				rep.Interrupted = true
				break
			}
			if done[rm+"|"+src] && !opts.Force {
				rep.Rows = append(rep.Rows, BackfillRow{Riksmote: rm, Source: src, Complete: true, Skipped: true})
				continue
			}
			var row BackfillRow
			if src == domain.BackfillMotions {
				row = b.doMotions(ctx, mw, rm)
			} else {
				row = b.doBetankanden(ctx, rm)
			}
			if row.Err != "" {
				rep.Failed++
				slog.Warn("backfill: riksmöte source failed", "rm", rm, "source", src, "error", row.Err)
			}
			rep.Rows = append(rep.Rows, row)
		}
	}

	if !opts.SkipRelink && !rep.Interrupted {
		n, err := b.svc.RelinkAll(ctx)
		rep.Relinked = n
		if err != nil {
			rep.Failed++
			slog.Warn("backfill: relink stopped", "relinked", n, "error", err)
		}
	}
	if err := b.fill(ctx, rep); err != nil {
		return rep, err
	}
	return rep, nil
}

// Report builds the report without fetching anything.
func (b *MotionsBackfill) Report(ctx context.Context) (*Report, error) {
	rep := &Report{}
	progress, err := b.svc.BackfillProgress(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range progress {
		rep.Rows = append(rep.Rows, BackfillRow{Riksmote: p.Riksmote, Source: p.Source,
			Expected: p.Expected, Held: p.Held, Complete: p.CompletedAt != nil})
	}
	return rep, b.fill(ctx, rep)
}

func (b *MotionsBackfill) fill(ctx context.Context, rep *Report) error {
	vocab, err := b.svc.DecisionVocabulary(ctx)
	if err != nil {
		return err
	}
	for _, v := range vocab {
		row := VocabularyRow{Field: v.Field, Value: v.Value, Count: v.Count}
		switch v.Field {
		case domain.FieldKammaren, domain.FieldPunkt:
			row.Mapped = domain.KnownOutcome(v.Value)
			row.Shown = domain.MappedOutcome(v.Value)
		default:
			// beslutstyp and vinnare are shown as Riksdagen words them.
			row.Mapped = true
			row.Shown = v.Value
			if v.Field == domain.FieldBeslutstyp {
				mode := domain.ModeOf(v.Value, "")
				row.Mapped = mode != domain.ModeAnnat
			}
		}
		rep.Vocabulary = append(rep.Vocabulary, row)
	}
	rep.Voteringar, rep.WithOrigin, err = b.svc.OriginCoverage(ctx)
	return err
}

func (b *MotionsBackfill) scope(ctx context.Context, opts BackfillOptions) ([]string, error) {
	var rms []string
	if len(opts.Riksmoten) > 0 {
		rms = append(rms, opts.Riksmoten...)
	} else {
		all, err := b.svc.BackfillRiksmoten(ctx)
		if err != nil {
			return nil, err
		}
		rms = all
	}
	from := opts.From
	if from == "" {
		from = OldestBackfillRiksmote
	}
	fromYear, _ := riksmoteStartYear(from)
	rms = OpenedRiksmoten(rms, b.at())
	out := rms[:0:0]
	for _, rm := range rms {
		if y, ok := riksmoteStartYear(rm); ok && y >= fromYear {
			out = append(out, rm)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] > out[j] })
	return out, nil
}

func (b *MotionsBackfill) doMotions(ctx context.Context, mw *MotionsWorker, rm string) BackfillRow {
	row := BackfillRow{Riksmote: rm, Source: domain.BackfillMotions}
	// An unbounded budget: the operator asked for the whole riksmöte, and a
	// restart re-lists and skips what is stored.
	if _, err := mw.SyncRiksmote(ctx, rm, 1<<30); err != nil {
		row.Err = err.Error()
	}
	cov, _ := b.svc.Coverage(ctx)
	for _, c := range cov {
		if c.Riksmote == rm && c.Source == domain.SourceRiksdagenDokumentlista {
			row.Expected = c.Expected
		}
	}
	held, err := b.svc.MotionCount(ctx, rm)
	if err != nil && row.Err == "" {
		row.Err = err.Error()
	}
	row.Held = held
	row.Complete = row.Err == "" && ctx.Err() == nil && row.Expected > 0 && held >= row.Expected
	b.save(ctx, row)
	return row
}

func (b *MotionsBackfill) doBetankanden(ctx context.Context, rm string) BackfillRow {
	row := BackfillRow{Riksmote: rm, Source: domain.BackfillBetankanden}
	tried := map[motionPorts.BetRef]bool{}
	for ctx.Err() == nil {
		refs, err := b.svc.Pending(ctx, rm, betankandeBatch)
		if err != nil {
			row.Err = err.Error()
			break
		}
		fresh := 0
		for _, ref := range refs {
			if tried[ref] {
				continue
			}
			tried[ref] = true
			fresh++
			if ctx.Err() != nil {
				break
			}
			if _, err := b.svc.SyncBetankande(ctx, ref); err != nil {
				row.Err = fmt.Sprintf("%s:%s: %v", ref.Riksmote, ref.Beteckning, err)
				slog.Warn("backfill: betänkande failed", "rm", ref.Riksmote, "bet", ref.Beteckning, "error", err)
			}
			time.Sleep(FetchDelay)
		}
		if fresh == 0 {
			break
		}
	}
	exp, held, err := b.svc.BetankandeCounts(ctx, rm)
	if err != nil && row.Err == "" {
		row.Err = err.Error()
	}
	row.Expected, row.Held = exp, held
	row.Complete = row.Err == "" && ctx.Err() == nil
	b.save(ctx, row)
	return row
}

func (b *MotionsBackfill) save(ctx context.Context, row BackfillRow) {
	p := domain.BackfillProgress{Riksmote: row.Riksmote, Source: row.Source, Expected: row.Expected, Held: row.Held}
	if row.Complete {
		t := time.Now()
		p.CompletedAt = &t
	}
	// The context may be cancelled; progress must still be written.
	if err := b.svc.SaveBackfillProgress(context.WithoutCancel(ctx), p); err != nil {
		slog.Warn("backfill: progress not saved", "rm", row.Riksmote, "source", row.Source, "error", err)
	}
}

// Write prints the report: coverage per riksmöte and source, every decision
// value seen with its count, the values Outcome does not map, and origins.
func (r *Report) Write(w io.Writer) {
	fmt.Fprintln(w, "== Täckning per riksmöte och källa ==")
	for _, row := range r.Rows {
		state := "ofullständig"
		switch {
		case row.Skipped:
			state = "klar (tidigare körning)"
		case row.Complete:
			state = "klar"
		}
		fmt.Fprintf(w, "%-8s %-12s förväntat %-6d har %-6d %s", row.Riksmote, row.Source, row.Expected, row.Held, state)
		if row.Err != "" {
			fmt.Fprintf(w, "  FEL: %s", row.Err)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "\n== Kammarens beslutsvärden (värde, antal, visas som) ==")
	var unmapped []VocabularyRow
	for _, v := range r.Vocabulary {
		mark := ""
		if !v.Mapped {
			mark = "  <-- EJ MAPPAT"
			unmapped = append(unmapped, v)
		}
		fmt.Fprintf(w, "%-12s %-40q %7d  -> %q%s\n", v.Field, v.Value, v.Count, v.Shown, mark)
	}
	fmt.Fprintln(w, "\n== Värden som Outcome inte mappar (visas ordagrant på sajten) ==")
	if len(unmapped) == 0 {
		fmt.Fprintln(w, "inga")
	}
	for _, v := range unmapped {
		fmt.Fprintf(w, "%s: %q (%d)\n", v.Field, v.Value, v.Count)
	}
	fmt.Fprintln(w, "\n== Ursprung ==")
	if r.Relink {
		fmt.Fprintf(w, "länkar omräknade för %d betänkanden\n", r.Relinked)
	}
	fmt.Fprintf(w, "voteringar med beslutade yrkanden länkade: %d av %d\n", r.WithOrigin, r.Voteringar)
	if r.Interrupted {
		fmt.Fprintln(w, "\nAVBRUTEN – starta om kommandot för att fortsätta där det slutade")
	}
	if r.Failed > 0 {
		fmt.Fprintf(w, "\n%d steg misslyckades – kör igen för att försöka på nytt\n", r.Failed)
	}
}
