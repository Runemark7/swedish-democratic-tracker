package ports

import (
	"context"
	"time"
)

// MandatePeriod is one parliament, bounded by the elections that open and
// close it. Riksmoten are the session labels Riksdagen files documents under.
type MandatePeriod struct {
	Code      string
	Start     time.Time
	End       time.Time
	Riksmoten []string
}

// MandatePeriodReader lists the periods the database has declared.
// Ingestion uses it so a new parliament is a row, not a code change.
type MandatePeriodReader interface {
	ListMandatePeriods(ctx context.Context) ([]MandatePeriod, error)
}
