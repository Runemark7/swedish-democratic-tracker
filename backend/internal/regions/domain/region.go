package domain

type Region struct {
	Code             string   `json:"code"`
	Name             string   `json:"name"`
	Capital          string   `json:"capital"`
	Population       int      `json:"population"`
	GoverningParties []string `json:"governingParties"`
	ElectionYear     int      `json:"electionYear"`
	TotalMandates    int      `json:"totalMandates"`
}

type RegionDetail struct {
	Region
	ElectionResults []ElectionResult `json:"electionResults"`
}

type Municipality struct {
	Code             string   `json:"code"`
	Name             string   `json:"name"`
	RegionCode       string   `json:"regionCode"`
	RegionName       string   `json:"regionName"`
	Population       int      `json:"population"`
	GoverningParties []string `json:"governingParties"`
	ElectionYear     int      `json:"electionYear"`
	TotalMandates    int      `json:"totalMandates"`
}

type MunicipalityDetail struct {
	Municipality
	ElectionResults []ElectionResult `json:"electionResults"`
}

// RegionPlan is a region's official annual-plan document plus the region's own
// overarching goals, transcribed verbatim. Goals is empty for regions whose
// plan has no cleanly extractable overarching set (link-only).
type RegionPlan struct {
	Code       string           `json:"code"`
	Label      string           `json:"label"`
	URL        string           `json:"url"`
	GoalsLabel string           `json:"goalsLabel,omitempty"`
	Period     string           `json:"period,omitempty"`
	Goals      []RegionPlanGoal `json:"goals"`
}

type RegionPlanGoal struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ElectionResult struct {
	Party         string  `json:"party"`
	Mandates      int     `json:"mandates"`
	VotePct       float64 `json:"votePct"`
	TotalMandates int     `json:"totalMandates"`
}
