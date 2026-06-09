package main

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		name     string
		ours     string
		upstream string
		volatile bool
		want     Status
	}{
		{"equal non-volatile", "283649", "283649", false, StatusMatch},
		{"differ non-volatile", "283649", "283700", false, StatusReview},
		{"equal volatile", "41233", "41233", true, StatusInfo},
		{"differ volatile", "41233", "41980", true, StatusInfo},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := classify(c.ours, c.upstream, c.volatile)
			if got != c.want {
				t.Fatalf("classify(%q,%q,%v) = %q, want %q", c.ours, c.upstream, c.volatile, got, c.want)
			}
		})
	}
}
