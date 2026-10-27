package workers

import (
	"fmt"
	"sort"
	"time"

	"riksdagskollen/internal/ingestion/ports"
)

// PreviousRiksmote is the session label one year before CurrentRiksmote.
// The September boundary is the same one CurrentRiksmote uses.
func PreviousRiksmote(t time.Time) string {
	y := t.Year()
	if t.Month() < time.September {
		y--
	}
	y--
	return fmt.Sprintf("%d/%02d", y, (y+1)%100)
}

// riksmoteStartYear parses the opening year of a "2026/27" label.
func riksmoteStartYear(rm string) (int, bool) {
	var year, suffix int
	if _, err := fmt.Sscanf(rm, "%d/%d", &year, &suffix); err != nil || year < 1900 {
		return 0, false
	}
	return year, true
}

// riksmoteAfter reports whether a opens after b. An unparseable label is not
// treated as future — dropping it would skip a session we failed to read.
func riksmoteAfter(a, b string) bool {
	ya, oka := riksmoteStartYear(a)
	yb, okb := riksmoteStartYear(b)
	if !oka || !okb {
		return false
	}
	return ya > yb
}

// OpenedRiksmoten keeps sessions that have reached the current riksmöte.
// Later labels in a mandate period are real, but Riksdagen has nothing in
// them yet, and asking every day only produces empty responses.
func OpenedRiksmoten(rms []string, now time.Time) []string {
	current := CurrentRiksmote(now)
	out := make([]string, 0, len(rms))
	for _, rm := range rms {
		if riksmoteAfter(rm, current) {
			continue
		}
		out = append(out, rm)
	}
	return out
}

// SpeechSessions is the daily speech fetch's scope.
//
// It is the riksmöten of the mandate period that has already opened (the one
// with the latest start date on or before now), excluding sessions that have
// not opened, plus two labels the table cannot be trusted to contain:
//
//   - the current riksmöte, so a missing row for the next parliament does not
//     freeze ingestion the way the hard-coded 2022–2026 list froze 2026/27
//   - the riksmöte that contains since, so the tail of the outgoing session
//     is still fetched after the September label rolls over
//
// since zero means there is no cursor. The previous riksmöte is included
// then, which is the same tail without a date to derive it from.
func SpeechSessions(periods []ports.MandatePeriod, now, since time.Time) []string {
	current := CurrentRiksmote(now)
	seen := map[string]bool{}
	var out []string
	add := func(rm string) {
		if rm == "" || seen[rm] || riksmoteAfter(rm, current) {
			return
		}
		seen[rm] = true
		out = append(out, rm)
	}

	if p := activeMandate(periods, now); p != nil {
		for _, rm := range p.Riksmoten {
			add(rm)
		}
	}
	add(current)
	if !since.IsZero() {
		add(CurrentRiksmote(since))
	} else {
		add(PreviousRiksmote(now))
	}

	sort.Slice(out, func(i, j int) bool {
		yi, oki := riksmoteStartYear(out[i])
		yj, okj := riksmoteStartYear(out[j])
		if oki && okj && yi != yj {
			return yi < yj
		}
		return out[i] < out[j]
	})
	return out
}

func activeMandate(periods []ports.MandatePeriod, now time.Time) *ports.MandatePeriod {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var best *ports.MandatePeriod
	var bestStart time.Time
	for i := range periods {
		p := &periods[i]
		start := time.Date(p.Start.Year(), p.Start.Month(), p.Start.Day(), 0, 0, 0, 0, time.UTC)
		if start.After(today) {
			continue
		}
		if best == nil || start.After(bestStart) {
			best = p
			bestStart = start
		}
	}
	return best
}
