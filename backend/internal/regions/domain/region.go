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

type ElectionResult struct {
	Party         string  `json:"party"`
	Mandates      int     `json:"mandates"`
	VotePct       float64 `json:"votePct"`
	TotalMandates int     `json:"totalMandates"`
}
