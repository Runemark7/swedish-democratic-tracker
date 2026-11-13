package postgres

import (
	"slices"
	"strings"
	"testing"
)

// Under the C locale Postgres folds ASCII only: upper('FöU') is 'FöU' and
// lower('FÖU3') is 'fÖu3'. Whatever the reader typed, the folded set must
// hold both what a UTF-8 locale and what the C locale make of the stored code.
func TestFoldAllCoversLocaleAwareAndASCIIOnlyFolding(t *testing.T) {
	for _, in := range []string{"FöU", "föu", "FÖU"} {
		got := upperAll([]string{in})
		if !slices.Contains(got, "FÖU") || !slices.Contains(got, "FöU") {
			t.Errorf("upperAll(%q) = %v", in, got)
		}
	}
	for _, in := range []string{"FöU3", "föu3", "FÖU3"} {
		got := foldAll([]string{in}, strings.ToLower, 'A', 'Z', 'a'-'A')
		if !slices.Contains(got, "föu3") || !slices.Contains(got, "fÖu3") {
			t.Errorf("lower fold of %q = %v", in, got)
		}
	}
	if got := upperAll([]string{"sku", "KU"}); !slices.Equal(got, []string{"SKU", "KU"}) {
		t.Errorf("ASCII codes fold to one value each: %v", got)
	}
}
