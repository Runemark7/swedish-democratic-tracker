package ports

import "testing"
import "time"

func boolPtr(v bool) *bool { return &v }

func TestDecided(t *testing.T) {
	now := time.Date(2026, 10, 27, 15, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		doc  RiksdagDocument
		want bool
	}{
		{
			name: "planned with a future beslutsdag",
			doc: RiksdagDocument{
				DecisionDate: "2026-12-09", Status: "planerat", ChamberDecided: boolPtr(false),
			},
		},
		{
			name: "published report scheduled for a future vote",
			doc: RiksdagDocument{
				DecisionDate: "2026-11-04", Status: "Webbpublicering", ChamberDecided: boolPtr(false),
			},
		},
		{
			name: "chamber has voted",
			doc: RiksdagDocument{
				DecisionDate: "2026-10-21", Status: "Webbpublicering", ChamberDecided: boolPtr(true),
			},
			want: true,
		},
		{
			name: "no flag, status planerat",
			doc:  RiksdagDocument{DecisionDate: "2026-12-09", Status: "planerat"},
		},
		{
			name: "no flag, decision date still in the future",
			doc:  RiksdagDocument{DecisionDate: "2026-12-09", Status: "Webbpublicering"},
		},
		{
			name: "no flag, decision date today",
			doc:  RiksdagDocument{DecisionDate: "2026-10-27", Status: "Webbpublicering"},
			want: true,
		},
		{
			name: "no flag and no date",
			doc:  RiksdagDocument{Status: "Webbpublicering"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.doc.Decided(now); got != tc.want {
				t.Errorf("Decided = %v, want %v", got, tc.want)
			}
		})
	}
}
