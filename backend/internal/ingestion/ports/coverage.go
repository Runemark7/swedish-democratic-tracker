package ports

import (
	"context"
	"time"
)

// Denominator is what Riksdagen's own listing says exists for one riksmöte.
//
// It is deliberately separate from how much we hold: the numerator is counted
// from our own rows at read time, so nothing here records it. Keeping a second
// copy of our row count in this table is what let the published figure freeze
// while the record kept growing.
type Denominator struct {
	Riksmote string
	// Expected is Riksdagen's @traffar for the riksmöte.
	Expected int
	// Unreachable are voteringar the listing includes but that carry no
	// beteckning, so the betänkande-partitioned fetch cannot reach them.
	Unreachable int
	// LastVoteDate is the newest decision date in the listing. It comes from the
	// listing's own `datum`, never from an ingest or update timestamp.
	LastVoteDate *time.Time
}

// CoverageRepository tracks the upstream side of the coverage claim.
type CoverageRepository interface {
	// RiksmotenNeedingDenominator lists the riksmöten whose denominator should
	// be re-read: every riksmöte of a mandate period that has not ended, plus
	// any whose denominator has never been established at all.
	//
	// Closed periods are otherwise left alone — once a parliament has finished
	// filing votes its denominator is final, so re-fetching it daily would buy
	// nothing.
	RiksmotenNeedingDenominator(ctx context.Context) ([]string, error)

	// UpsertDenominator records a denominator and stamps it with the time it was
	// read, so its age can be stated rather than implied.
	UpsertDenominator(ctx context.Context, d Denominator) error
}
