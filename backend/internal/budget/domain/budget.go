package domain

import (
	"fmt"
	"net/url"
)

type ExpenditureArea struct {
	ID          int    `json:"id,omitempty"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sortOrder"`
}

type BudgetYear struct {
	Year     int    `json:"year"`
	Status   string `json:"status"`
	TotalKsek int64 `json:"totalKsek"`
}

type BudgetAllocation struct {
	Area            ExpenditureArea `json:"area"`
	AmountKsek      int64           `json:"amountKsek"`
	AmountFormatted string          `json:"amountFormatted"`
}

type BudgetYearDetail struct {
	Year        int                 `json:"year"`
	Status      string              `json:"status"`
	TotalKsek   int64               `json:"totalKsek"`
	Allocations []BudgetAllocation  `json:"allocations"`
	Documents   []BudgetDocumentRef `json:"documents,omitempty"`
}

type ComparisonRow struct {
	Area             ExpenditureArea `json:"area"`
	BaseAmountKsek   int64           `json:"baseAmountKsek"`
	CompareAmountKsek int64          `json:"compareAmountKsek"`
	DeltaKsek        int64           `json:"deltaKsek"`
	DeltaPct         float64         `json:"deltaPct"`
}

type BudgetComparison struct {
	BaseYear        int             `json:"baseYear"`
	CompareYear     int             `json:"compareYear"`
	BaseTotalKsek   int64           `json:"baseTotalKsek"`
	CompareTotalKsek int64          `json:"compareTotalKsek"`
	TotalDeltaKsek  int64           `json:"totalDeltaKsek"`
	TotalDeltaPct   float64         `json:"totalDeltaPct"`
	Rows            []ComparisonRow `json:"rows"`
}

type AreaTimeSeriesEntry struct {
	Year       int                `json:"year"`
	AmountKsek int64              `json:"amountKsek"`
	Documents  []BudgetDocumentRef `json:"documents,omitempty"`
}

// BudgetDocumentRef links to a Riksdagen or government document
// that explains a budget decision.
type BudgetDocumentRef struct {
	Type        string `json:"type"`        // "proposition", "committee_report"
	Title       string `json:"title"`       // e.g. "Prop. 2024/25:1"
	URL         string `json:"url"`         // link to riksdagen.se or data.riksdagen.se
	Description string `json:"description"` // e.g. "Budgetpropositionen för 2025"
}

type AreaTimeSeries struct {
	Area    ExpenditureArea       `json:"area"`
	Entries []AreaTimeSeriesEntry `json:"entries"`
}

// FormatKsek formats KSEK to a human-readable Swedish budget string.
// e.g. 145200000 KSEK -> "145,2 mdkr"
func FormatKsek(ksek int64) string {
	mdkr := float64(ksek) / 1_000_000 // KSEK -> mdkr
	if mdkr >= 1 || mdkr <= -1 {
		// Use one decimal
		s := formatFloat(mdkr, 1)
		return s + " mdkr"
	}
	mnkr := float64(ksek) / 1_000 // KSEK -> mnkr
	s := formatFloat(mnkr, 0)
	return s + " mnkr"
}

func formatFloat(f float64, decimals int) string {
	negative := f < 0
	if negative {
		f = -f
	}
	// Format with Swedish comma as decimal separator
	intPart := int64(f)
	var s string
	if decimals > 0 {
		frac := f - float64(intPart)
		for range decimals {
			frac *= 10
		}
		fracInt := int64(frac + 0.5)
		format := "%d,%0" + string(rune('0'+decimals)) + "d"
		_ = format
		// Simple approach
		if decimals == 1 {
			s = intToString(intPart) + "," + intToString(fracInt)
		} else {
			s = intToString(intPart)
		}
	} else {
		s = intToString(int64(f + 0.5))
	}
	if negative {
		return "-" + s
	}
	return s
}

// budgetSession returns the riksmöte session string for a budget year.
// Budget for year 2025 → proposed in session "2024/25".
func budgetSession(year int) string {
	return fmt.Sprintf("%d/%02d", year-1, year%100)
}

// riksmoteCode returns the Riksdag two-character session code.
// H1=2013/14, H2=2014/15, ..., H9=2021/22, HA=2022/23, HB=2023/24, etc.
// Returns empty string for years outside the supported range.
func riksmoteCode(budgetYear int) string {
	offset := budgetYear - 2013
	if offset < 1 || offset > 35 {
		return ""
	}
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return "H" + string(chars[offset])
}

// DocumentRefsForYear returns links to the budget proposition and finance
// committee reports for a given budget year. Pure computation — no API calls.
func DocumentRefsForYear(year int) []BudgetDocumentRef {
	session := budgetSession(year)
	code := riksmoteCode(year)

	refs := []BudgetDocumentRef{
		{
			Type:        "proposition",
			Title:       fmt.Sprintf("Prop. %s:1", session),
			URL:         fmt.Sprintf("https://data.riksdagen.se/dokumentlista/?doktyp=prop&rm=%s&bet=1&utformat=html&sz=50", url.QueryEscape(session)),
			Description: fmt.Sprintf("Budgetpropositionen för %d", year),
		},
	}

	// Add direct link to main proposition if code is known
	if code != "" {
		refs[0].URL = fmt.Sprintf("https://data.riksdagen.se/dokument/%s031.html", code)
	}

	// Finance Committee report search
	refs = append(refs, BudgetDocumentRef{
		Type:        "committee_report",
		Title:       fmt.Sprintf("FiU — %s", session),
		URL:         fmt.Sprintf("https://data.riksdagen.se/dokumentlista/?doktyp=bet&rm=%s&org=FiU&utformat=html&sz=50", url.QueryEscape(session)),
		Description: fmt.Sprintf("Finansutskottets betänkanden %s", session),
	})

	return refs
}

// DocumentRefsForArea returns links to the specific utgiftsområde volume
// of the budget proposition for a given year and area number (1-27).
func DocumentRefsForArea(year, areaNumber int) []BudgetDocumentRef {
	session := budgetSession(year)
	code := riksmoteCode(year)

	if code == "" {
		// Fall back to search URL
		return []BudgetDocumentRef{
			{
				Type:        "proposition",
				Title:       fmt.Sprintf("Prop. %s:1 UO%d", session, areaNumber),
				URL:         fmt.Sprintf("https://data.riksdagen.se/dokumentlista/?doktyp=prop&rm=%s&bet=1&utformat=html&sz=50", url.QueryEscape(session)),
				Description: fmt.Sprintf("Budgetpropositionen för %d — Utgiftsområde %d", year, areaNumber),
			},
		}
	}

	// Direct link: d2=UO1, d3=UO2, ..., d28=UO27
	dokID := fmt.Sprintf("%s031d%d", code, areaNumber+1)
	return []BudgetDocumentRef{
		{
			Type:        "proposition",
			Title:       fmt.Sprintf("Prop. %s:1 UO%d", session, areaNumber),
			URL:         fmt.Sprintf("https://data.riksdagen.se/dokument/%s.html", dokID),
			Description: fmt.Sprintf("Budgetpropositionen för %d — Utgiftsområde %d", year, areaNumber),
		},
	}
}

func intToString(n int64) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	digits := make([]byte, 0, 20)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}
