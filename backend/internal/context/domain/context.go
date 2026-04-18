package domain

import "fmt"

// BudgetTrendEntry is one year's allocation for a budget area.
type BudgetTrendEntry struct {
	Year            int    `json:"year"`
	AmountKsek      int64  `json:"amountKsek"`
	AmountFormatted string `json:"amountFormatted"`
}

// RelatedBudgetArea is a budget area relevant to a topic, with multi-year trend.
type RelatedBudgetArea struct {
	Code             string             `json:"code"`
	Name             string             `json:"name"`
	Trend            []BudgetTrendEntry `json:"trend"`
	LatestDeltaPct   float64            `json:"latestDeltaPct"`
	ShareOfBudgetPct float64            `json:"shareOfBudgetPct"`
}

// FundingRow is one budget area's year-over-year delta, used in funding analysis.
type FundingRow struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	DeltaKsek int64   `json:"deltaKsek"`
	DeltaPct  float64 `json:"deltaPct"`
}

// FundingContext shows which areas increased and decreased in the latest budget year.
type FundingContext struct {
	Year                 int          `json:"year"`
	TopIncreases         []FundingRow `json:"topIncreases"`
	TopDecreases         []FundingRow `json:"topDecreases"`
	TotalBudgetDeltaKsek int64        `json:"totalBudgetDeltaKsek"`
	TotalBudgetDeltaPct  float64      `json:"totalBudgetDeltaPct"`
}

// TopicContext is the full consequence context for a political topic.
type TopicContext struct {
	Topic          string              `json:"topic"`
	Description    string              `json:"description"`
	Committees     []string            `json:"committees"`
	RelatedAreas   []RelatedBudgetArea `json:"relatedAreas"`
	FundingContext *FundingContext      `json:"fundingContext,omitempty"`
}

// FormatKsek formats a KSEK amount to Swedish budget notation (e.g. "87,3 mdkr").
func FormatKsek(ksek int64) string {
	if ksek == 0 {
		return "0 mdkr"
	}
	mdkr := float64(ksek) / 1_000_000
	if mdkr >= 1 || mdkr <= -1 {
		return fmt.Sprintf("%.1f mdkr", mdkr)
	}
	mnkr := float64(ksek) / 1_000
	return fmt.Sprintf("%.0f mnkr", mnkr)
}
