package ingestion

import (
	"strings"
	"testing"
	"time"
)

func TestMergeCatalog_KeepsNeverRunWorkers(t *testing.T) {
	got := MergeCatalog(
		[]ScheduledJob{{Name: "fetch-speeches", Spec: "@daily"}, {Name: "refresh-scorecards", Spec: "@weekly"}},
		[]JobFreshness{{Name: "fetch-speeches", LastStatus: "success"}},
	)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Spec != "@daily" || got[0].LastStatus != "success" {
		t.Errorf("speeches row = %+v", got[0])
	}
	if got[1].Name != "refresh-scorecards" || got[1].Spec != "@weekly" || got[1].LastSuccessAt != nil {
		t.Errorf("scorecards row = %+v", got[1])
	}
}

func TestAlerts_ContentAndJobs(t *testing.T) {
	now := time.Date(2026, 10, 27, 6, 0, 0, 0, time.UTC)
	oldSpeech := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	freshVote := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	recent := now.Add(-2 * time.Hour)
	staleDaily := now.Add(-48 * time.Hour)
	weekAgo := now.Add(-7 * 24 * time.Hour)

	alerts := Alerts(Snapshot{
		SpeechesLatest: &oldSpeech,
		VotesLatest:    &freshVote,
		Jobs: []JobFreshness{
			{Name: "fetch-speeches", Spec: "@daily", LastSuccessAt: &recent},
			{Name: "fetch-votes", Spec: "@daily", LastSuccessAt: &staleDaily},
			{Name: "refresh-scorecards", Spec: "@weekly", LastSuccessAt: &weekAgo},
			{Name: "fetch-politicians", Spec: "@daily"},
		},
	}, now)

	want := []string{
		"speeches: newest record 2026-08-13 is 75d old (threshold 60d)",
		"fetch-votes: last success 2026-10-25T06:00:00Z is 48h old (daily threshold 36h)",
		"fetch-politicians: has never succeeded",
	}
	if len(alerts) != len(want) {
		t.Fatalf("alerts = %v, want %v", alerts, want)
	}
	for i := range want {
		if alerts[i] != want[i] {
			t.Errorf("alerts[%d] = %q, want %q", i, alerts[i], want[i])
		}
	}
}

func TestAlerts_QuietWhenDataAndJobsAreRecent(t *testing.T) {
	now := time.Date(2026, 10, 27, 6, 0, 0, 0, time.UTC)
	speech := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
	vote := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	ran := now.Add(-time.Hour)
	alerts := Alerts(Snapshot{
		SpeechesLatest: &speech,
		VotesLatest:    &vote,
		Jobs:           []JobFreshness{{Name: "fetch-speeches", Spec: "@daily", LastSuccessAt: &ran}},
	}, now)
	if len(alerts) != 0 {
		t.Fatalf("alerts = %v, want none", alerts)
	}
}

// The freshness report and this alert must name the stored calendar day.
// A speech dated 2026-10-21 is midnight UTC once the DATE is scanned without
// a session-timezone cast. Age is measured from that same midnight, which is
// how the GitHub check compares the date strings.
func TestAlerts_NamesTheStoredSpeechDay(t *testing.T) {
	now := time.Date(2026, 12, 27, 6, 0, 0, 0, time.UTC)
	speech := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
	alerts := Alerts(Snapshot{SpeechesLatest: &speech}, now)
	var speechAlert string
	for _, alert := range alerts {
		if strings.HasPrefix(alert, "speeches:") {
			speechAlert = alert
		}
	}
	if !strings.Contains(speechAlert, "2026-10-21") || strings.Contains(speechAlert, "2026-10-20") {
		t.Fatalf("speech alert = %q, want the stored day 2026-10-21", speechAlert)
	}
}

// Midnight of a DATE in Europe/Stockholm is the previous evening in UTC
// (CEST on 2026-10-21, so 22:00Z). Formatting that instant prints 2026-10-20.
// The repository must not hand this instant to the alert or the JSON report.
func TestAlerts_LocalMidnightPrintsThePreviousUTCDay(t *testing.T) {
	stockholm, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		t.Fatal(err)
	}
	cast := time.Date(2026, 10, 21, 0, 0, 0, 0, stockholm)
	now := time.Date(2026, 12, 27, 6, 0, 0, 0, time.UTC)
	alerts := Alerts(Snapshot{SpeechesLatest: &cast}, now)
	var speechAlert string
	for _, alert := range alerts {
		if strings.HasPrefix(alert, "speeches:") {
			speechAlert = alert
		}
	}
	if !strings.Contains(speechAlert, "2026-10-20") {
		t.Fatalf("speech alert = %q, want 2026-10-20 from the timestamptz cast", speechAlert)
	}
}
