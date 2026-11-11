// Package domain holds the Data section's coverage vocabulary.
package domain

import "time"

// Source ids as the Data section names them.
const (
	SourceMotioner         = "motioner"
	SourceBetankanden      = "betankanden"
	SourceVoteringar       = "voteringar"
	SourceAnforanden       = "anforanden"
	SourceFragor           = "fragor"
	SourceInterpellationer = "interpellationer"
)

// Status of one source in one riksmöte. A riksmöte nobody has read has no row
// at all: absence is shown as "not loaded", never as zero.
const (
	StatusComplete = "complete"
	StatusPartial  = "partial"
)

// Row is what we hold of one source in one riksmöte.
type Row struct {
	Source     string
	Riksmote   string
	Status     string
	Expected   int
	Held       int
	LastUpdate time.Time
}
