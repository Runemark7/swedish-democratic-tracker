package workers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"time"

	"riksdagskollen/internal/politicians"
	"riksdagskollen/internal/speeches"
	"riksdagskollen/internal/speeches/domain"
	speechPorts "riksdagskollen/internal/speeches/ports"
)

// OldestSpeechBackfillRiksmote is where the first speech backfill starts
// (issue #147: the two newest mandate periods and 2014–2018). Older periods
// follow in #148.
const OldestSpeechBackfillRiksmote = "2014/15"

// BackfillParties are the party codes the speech list is asked for. The list
// cannot be paged, so a riksmöte is read one party at a time. FP is
// Folkpartiet, whose code was FP until it became L in 2015. Speakers without a
// party (talmän, ministers outside the Riksdag) are not in any party's list.
var BackfillParties = append(append([]string{}, ActiveParties...), "FP")

// speechTextBatch is how many speeches without text are read at a time.
const speechTextBatch = 200

// SpeechBackfillOptions steers one run of the speech backfill.
type SpeechBackfillOptions struct {
	// Riksmoten limits the run to these. Empty means From..To.
	Riksmoten []string
	// From is the oldest riksmöte to take, To the newest (default: the
	// current one). Both are labels like "2017/18".
	From, To string
	// Force redoes riksmöten recorded as complete.
	Force bool
}

// SpeechesBackfill is the operator's one-riksmöte-at-a-time ingestion of
// anföranden. It is never registered with the scheduler.
//
// Resumable: progress is recorded per riksmöte, and a restart skips what is
// complete and redoes the rest. Idempotent: every write is an upsert and a
// re-listed speech keeps its text. Polite: it pauses FetchDelay between
// requests.
type SpeechesBackfill struct {
	speeches    *speeches.Service
	politicians *politicians.Service
	now         func() time.Time
}

func NewSpeechesBackfill(sp *speeches.Service, pol *politicians.Service) *SpeechesBackfill {
	return &SpeechesBackfill{speeches: sp, politicians: pol}
}

// SetNow pins the clock that decides which riksmöten have opened.
func (b *SpeechesBackfill) SetNow(now func() time.Time) { b.now = now }

func (b *SpeechesBackfill) at() time.Time {
	if b.now != nil {
		return b.now().UTC()
	}
	return time.Now().UTC()
}

// SpeechBackfillRow is one riksmöte of a run.
type SpeechBackfillRow struct {
	Riksmote string
	Expected int // distinct speeches the lists report
	Stored   int
	Linked   int
	Unlinked int
	// MissingText counts stored speeches whose text could not be read.
	MissingText        int
	TextBytes          int64
	RowsAdded          int
	TableBytesAdded    int64
	IndexBytesAdded    int64
	PoliticiansCreated int
	Unresolved         []string // speakers Riksdagen has no person for
	Complete           bool
	Skipped            bool // already complete from an earlier run
	Err                string
}

// SpeechReport is what a run found.
type SpeechReport struct {
	Rows        []SpeechBackfillRow
	Failed      int
	Interrupted bool
}

// Riksmoten lists the labels from..to, newest first.
func SpeechBackfillRiksmoten(from, to string) []string {
	fy, ok1 := riksmoteStartYear(from)
	ty, ok2 := riksmoteStartYear(to)
	if !ok1 || !ok2 || fy > ty {
		return nil
	}
	var out []string
	for y := ty; y >= fy; y-- {
		out = append(out, fmt.Sprintf("%d/%02d", y, (y+1)%100))
	}
	return out
}

func (b *SpeechesBackfill) scope(opts SpeechBackfillOptions) []string {
	if len(opts.Riksmoten) > 0 {
		rms := OpenedRiksmoten(opts.Riksmoten, b.at())
		sort.Slice(rms, func(i, j int) bool { return rms[i] > rms[j] })
		return rms
	}
	from, to := opts.From, opts.To
	if from == "" {
		from = OldestSpeechBackfillRiksmote
	}
	if cur := CurrentRiksmote(b.at()); to == "" || riksmoteAfter(to, cur) {
		to = cur
	}
	return SpeechBackfillRiksmoten(from, to)
}

// Run walks the riksmöten newest first. It returns the report even when the
// context is cancelled part-way; Interrupted says so.
func (b *SpeechesBackfill) Run(ctx context.Context, opts SpeechBackfillOptions) (*SpeechReport, error) {
	progress, err := b.speeches.ListBackfill(ctx)
	if err != nil {
		return nil, err
	}
	done := map[string]bool{}
	for _, p := range progress {
		if p.CompletedAt != nil {
			done[p.Riksmote] = true
		}
	}
	rep := &SpeechReport{}
	for _, rm := range b.scope(opts) {
		if ctx.Err() != nil {
			rep.Interrupted = true
			break
		}
		if done[rm] && !opts.Force {
			rep.Rows = append(rep.Rows, SpeechBackfillRow{Riksmote: rm, Complete: true, Skipped: true})
			continue
		}
		row := b.doRiksmote(ctx, rm)
		if row.Err != "" || !row.Complete {
			rep.Failed++
			slog.Warn("speech backfill: riksmöte not complete", "rm", rm, "error", row.Err)
		}
		rep.Rows = append(rep.Rows, row)
		if ctx.Err() != nil {
			rep.Interrupted = true
			break
		}
	}
	return rep, nil
}

func (b *SpeechesBackfill) doRiksmote(ctx context.Context, rm string) SpeechBackfillRow {
	row := SpeechBackfillRow{Riksmote: rm}
	fail := func(err error) {
		if row.Err == "" {
			row.Err = err.Error()
		}
	}
	tableBefore, indexBefore, err := b.speeches.StorageSizes(ctx)
	if err != nil {
		fail(err)
		return row
	}
	before, err := b.speeches.RiksmoteCounts(ctx, rm)
	if err != nil {
		fail(err)
		return row
	}

	// 1. The lists, one party at a time.
	type key struct{ dok, nr string }
	seen := map[key]bool{}
	var listed []*domain.Speech
	listingOK := true
	for _, party := range BackfillParties {
		if ctx.Err() != nil {
			listingOK = false
			break
		}
		ss, err := b.speeches.Fetch(ctx, speechPorts.FetchSpeechesFilter{Session: rm, Party: party, Size: SpeechFetchLimit})
		if err != nil {
			listingOK = false
			fail(fmt.Errorf("%s: %w", party, err))
			slog.Warn("speech backfill: list failed", "rm", rm, "party", party, "error", err)
		} else if len(ss) >= SpeechFetchLimit {
			listingOK = false
			fail(fmt.Errorf("%s: the list returned the fetch limit (%d) and cannot be paged", party, SpeechFetchLimit))
		} else {
			for _, s := range ss {
				if s.Riksmote == "" {
					s.Riksmote = rm
				}
				k := key{s.DokID, s.AnforandeNummer}
				if !seen[k] {
					seen[k] = true
					listed = append(listed, s)
				}
			}
		}
		time.Sleep(FetchDelay)
	}
	row.Expected = len(listed)

	// 2. Every speaker gets a politicians row before a speech is stored.
	idSet := map[string]bool{}
	for _, s := range listed {
		idSet[s.PoliticianID] = true
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	created, unresolved, err := b.politicians.CreateMissing(ctx, ids)
	if err != nil {
		fail(err)
	}
	row.PoliticiansCreated, row.Unresolved = created, unresolved

	// 3. Store. Speeches of speakers Riksdagen has no person for stay out.
	if err == nil {
		skip := map[string]bool{}
		for _, id := range unresolved {
			skip[id] = true
		}
		toStore := listed[:0:0]
		for _, s := range listed {
			if !skip[s.PoliticianID] {
				toStore = append(toStore, s)
			}
		}
		for start := 0; start < len(toStore) && ctx.Err() == nil; start += 1000 {
			end := min(start+1000, len(toStore))
			if err := b.speeches.Store(ctx, toStore[start:end]); err != nil {
				fail(err)
				break
			}
		}
	}

	// 4. The text of every speech that has none yet.
	b.fillText(ctx, rm, fail)

	// 5. Link through #146's rules.
	if ctx.Err() == nil {
		if _, err := b.speeches.RelinkDebatesIn(ctx, rm); err != nil {
			fail(err)
		}
	}

	// 6. Measure and record. The context may be cancelled; progress must still be written.
	wctx := context.WithoutCancel(ctx)
	after, err := b.speeches.RiksmoteCounts(wctx, rm)
	if err != nil {
		fail(err)
	}
	tableAfter, indexAfter, err := b.speeches.StorageSizes(wctx)
	if err != nil {
		fail(err)
	}
	row.Stored, row.Linked, row.Unlinked = after.Stored, after.Linked, after.Unlinked()
	row.MissingText, row.TextBytes = after.MissingText, after.TextBytes
	row.RowsAdded = after.Stored - before.Stored
	row.TableBytesAdded, row.IndexBytesAdded = tableAfter-tableBefore, indexAfter-indexBefore
	row.Complete = row.Err == "" && listingOK && ctx.Err() == nil && len(unresolved) == 0 &&
		row.Expected > 0 && after.Stored >= row.Expected && after.MissingText == 0
	if err := b.speeches.SaveBackfill(wctx, domain.BackfillRun{
		Riksmote: rm, Expected: row.Expected, TextBytes: row.TextBytes, RowsAdded: row.RowsAdded,
		TableBytesAdded: row.TableBytesAdded, IndexBytesAdded: row.IndexBytesAdded,
		PoliticiansCreated: row.PoliticiansCreated, Unresolved: len(unresolved), Complete: row.Complete,
	}); err != nil {
		fail(err)
		row.Complete = false
	}
	return row
}

// fillText reads the body of every stored speech of the riksmöte that has none.
// A speech whose read fails is tried once per run and left for the next.
func (b *SpeechesBackfill) fillText(ctx context.Context, rm string, fail func(error)) {
	tried := map[int]bool{}
	for ctx.Err() == nil {
		batch, err := b.speeches.MissingTextIn(ctx, rm, speechTextBatch+len(tried))
		if err != nil {
			fail(err)
			return
		}
		fresh := 0
		for _, sp := range batch {
			if tried[sp.ID] {
				continue
			}
			tried[sp.ID] = true
			fresh++
			if ctx.Err() != nil {
				return
			}
			if err := b.speeches.EnrichOne(ctx, sp); err != nil {
				fail(fmt.Errorf("speech %s-%s: %w", sp.DokID, sp.AnforandeNummer, err))
				slog.Warn("speech backfill: text failed", "dok", sp.DokID, "nr", sp.AnforandeNummer, "error", err)
			}
			time.Sleep(FetchDelay)
		}
		if fresh == 0 {
			return
		}
	}
}

const mib = 1 << 20

// Write prints the report: coverage and the storage growth per riksmöte, and
// the total with a per-riksmöte average to size the older periods by.
func (r *SpeechReport) Write(w io.Writer) {
	fmt.Fprintln(w, "== Täckning per riksmöte ==")
	var rows int
	var text, table, index int64
	worked := 0
	for _, row := range r.Rows {
		state := "ofullständig"
		switch {
		case row.Skipped:
			state = "klar (tidigare körning)"
		case row.Complete:
			state = "klar"
		}
		if row.Skipped {
			fmt.Fprintf(w, "%-8s %s\n", row.Riksmote, state)
			continue
		}
		fmt.Fprintf(w, "%-8s förväntat %-6d lagrade %-6d kopplade %-6d okopplade %-6d saknar text %-4d %s",
			row.Riksmote, row.Expected, row.Stored, row.Linked, row.Unlinked, row.MissingText, state)
		if row.PoliticiansCreated > 0 {
			fmt.Fprintf(w, "  nya ledamöter %d", row.PoliticiansCreated)
		}
		if len(row.Unresolved) > 0 {
			fmt.Fprintf(w, "  OKÄNDA TALARE %v", row.Unresolved)
		}
		if row.Err != "" {
			fmt.Fprintf(w, "  FEL: %s", row.Err)
		}
		fmt.Fprintln(w)
		worked++
		rows += row.RowsAdded
		text += row.TextBytes
		table += row.TableBytesAdded
		index += row.IndexBytesAdded
	}
	fmt.Fprintln(w, "\n== Lagringstillväxt per riksmöte (nya rader, textbyte, tabell, index) ==")
	for _, row := range r.Rows {
		if row.Skipped {
			continue
		}
		fmt.Fprintf(w, "%-8s +%d rader  text %.1f MiB  tabell %+.1f MiB  index %+.1f MiB\n", row.Riksmote,
			row.RowsAdded, float64(row.TextBytes)/mib, float64(row.TableBytesAdded)/mib, float64(row.IndexBytesAdded)/mib)
	}
	if worked > 0 {
		fmt.Fprintf(w, "summa    +%d rader  text %.1f MiB  tabell %+.1f MiB  index %+.1f MiB  (snitt per riksmöte: %d rader, %.1f MiB tabell+index)\n",
			rows, float64(text)/mib, float64(table)/mib, float64(index)/mib, rows/worked, float64(table+index)/float64(worked)/mib)
	}
	if r.Interrupted {
		fmt.Fprintln(w, "\nAVBRUTEN – starta om kommandot för att fortsätta där det slutade")
	}
	if r.Failed > 0 {
		fmt.Fprintf(w, "\n%d riksmöten är inte klara – kör igen för att försöka på nytt\n", r.Failed)
	}
}
