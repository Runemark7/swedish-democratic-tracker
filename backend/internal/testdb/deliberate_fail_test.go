package testdb

import "testing"

// Proves a failing Go test turns the CI check red. This branch is not for merge.
func TestDeliberateCIFailure(t *testing.T) {
	t.Fatal("deliberate failure to prove a failing Go test turns the CI check red")
}
