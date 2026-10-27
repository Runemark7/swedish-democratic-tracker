package workers

import (
	"slices"
	"testing"
	"time"

	"riksdagskollen/internal/ingestion/ports"
)

func TestPreviousRiksmote(t *testing.T) {
	cases := map[string]string{
		"2026-08-01": "2024/25",
		"2026-09-20": "2025/26",
		"2026-10-27": "2025/26",
		"2030-10-15": "2029/30",
	}
	for in, want := range cases {
		d, err := time.Parse("2006-01-02", in)
		if err != nil {
			t.Fatal(err)
		}
		if got := PreviousRiksmote(d); got != want {
			t.Errorf("PreviousRiksmote(%s) = %q, want %q", in, got, want)
		}
	}
}

func period2022() ports.MandatePeriod {
	return ports.MandatePeriod{
		Code:      "2022-2026",
		Start:     time.Date(2022, 9, 11, 0, 0, 0, 0, time.UTC),
		End:       time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
		Riksmoten: []string{"2022/23", "2023/24", "2024/25", "2025/26"},
	}
}

func period2026() ports.MandatePeriod {
	return ports.MandatePeriod{
		Code:      "2026-2030",
		Start:     time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
		End:       time.Date(2030, 9, 8, 0, 0, 0, 0, time.UTC),
		Riksmoten: []string{"2026/27", "2027/28", "2028/29", "2029/30"},
	}
}

func TestSpeechSessions_CurrentPeriodAndCursorTail(t *testing.T) {
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	since := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	got := SpeechSessions([]ports.MandatePeriod{period2022(), period2026()}, now, since)
	want := []string{"2025/26", "2026/27"}
	if !slices.Equal(got, want) {
		t.Fatalf("SpeechSessions = %v, want %v", got, want)
	}
}

func TestSpeechSessions_DropsFutureSessionsAndOldOnesOnceCursorMoves(t *testing.T) {
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	since := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
	got := SpeechSessions([]ports.MandatePeriod{period2026()}, now, since)
	want := []string{"2026/27"}
	if !slices.Equal(got, want) {
		t.Fatalf("SpeechSessions = %v, want %v", got, want)
	}
}

func TestSpeechSessions_BeforeTheNewPeriodUsesTheOpenOne(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	since := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	got := SpeechSessions([]ports.MandatePeriod{period2022(), period2026()}, now, since)
	want := []string{"2022/23", "2023/24", "2024/25", "2025/26"}
	if !slices.Equal(got, want) {
		t.Fatalf("SpeechSessions = %v, want %v", got, want)
	}
}

func TestSpeechSessions_IncludesCurrentRiksmoteWhenTheTableLags(t *testing.T) {
	// October 2030 is riksmöte 2030/31. The newest row still ends at 2029/30.
	// The hard-coded list froze exactly this way in 2026.
	now := time.Date(2030, 10, 15, 0, 0, 0, 0, time.UTC)
	since := time.Date(2030, 8, 13, 0, 0, 0, 0, time.UTC)
	got := SpeechSessions([]ports.MandatePeriod{period2026()}, now, since)
	for _, rm := range []string{"2029/30", "2030/31"} {
		if !slices.Contains(got, rm) {
			t.Errorf("sessions %v missing %s", got, rm)
		}
	}
	if slices.Contains(got, "2031/32") {
		t.Errorf("sessions %v include a riksmöte that has not opened", got)
	}
}

func TestSpeechSessions_EmptyTableStillFollowsTheCalendar(t *testing.T) {
	now := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	since := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	got := SpeechSessions(nil, now, since)
	want := []string{"2025/26", "2026/27"}
	if !slices.Equal(got, want) {
		t.Fatalf("SpeechSessions = %v, want %v", got, want)
	}
}

func TestOpenedRiksmoten_SkipsSessionsThatHaveNotOpened(t *testing.T) {
	now := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	got := OpenedRiksmoten([]string{"2025/26", "2026/27", "2027/28", "not-a-session"}, now)
	want := []string{"2025/26", "2026/27", "not-a-session"}
	if !slices.Equal(got, want) {
		t.Fatalf("OpenedRiksmoten = %v, want %v", got, want)
	}
}
