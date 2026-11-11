package domain

import "time"

// RiksmoteCounts is what is stored for one riksmöte, read live.
type RiksmoteCounts struct {
	Stored      int
	Linked      int // linked to a betänkande
	MissingText int
	TextBytes   int64
}

// Unlinked is the stored speeches no betänkande could be named for.
func (c RiksmoteCounts) Unlinked() int { return c.Stored - c.Linked }

// BackfillRun is one run's work on a riksmöte, saved to speech_backfill. The
// *Added fields are added to what earlier runs recorded.
type BackfillRun struct {
	Riksmote           string
	Expected           int
	TextBytes          int64
	RowsAdded          int
	TableBytesAdded    int64
	IndexBytesAdded    int64
	PoliticiansCreated int
	Unresolved         int
	// Complete sets completed_at; false clears it.
	Complete bool
}

// Backfill is a riksmöte the backfill has worked on: what it recorded, with
// the counts of speeches read live.
type Backfill struct {
	Riksmote           string
	Expected           int
	Stored             int
	Linked             int
	Unlinked           int
	TextBytes          int64
	RowsAdded          int
	TableBytesAdded    int64
	IndexBytesAdded    int64
	PoliticiansCreated int
	Unresolved         int
	CompletedAt        *time.Time
}
