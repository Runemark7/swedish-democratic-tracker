package domain

// displayNames gives each committee code its Swedish name.
//
// This is a display aid and never a gate. A code absent here still exists in
// the record, so DisplayName falls back to the code itself: "the 15 standing
// committees" is a fact about the present, not about the record, which also
// holds joint committees and ones abolished in 2006.
var displayNames = map[string]string{
	"AU":   "Arbetsmarknadsutskottet",
	"CU":   "Civilutskottet",
	"FiU":  "Finansutskottet",
	"FöU":  "Försvarsutskottet",
	"JuU":  "Justitieutskottet",
	"KrU":  "Kulturutskottet",
	"KU":   "Konstitutionsutskottet",
	"MJU":  "Miljö- och jordbruksutskottet",
	"NU":   "Näringsutskottet",
	"SfU":  "Socialförsäkringsutskottet",
	"SkU":  "Skatteutskottet",
	"SoU":  "Socialutskottet",
	"TU":   "Trafikutskottet",
	"UbU":  "Utbildningsutskottet",
	"UU":   "Utrikesutskottet",
	"UFöU": "Sammansatta utrikes- och försvarsutskottet",
	"LU":   "Lagutskottet",
	"BoU":  "Bostadsutskottet",
}

// DisplayName returns the committee's Swedish name, or the code itself when we
// have no name for it.
func DisplayName(code string) string {
	if name, ok := displayNames[code]; ok {
		return name
	}
	return code
}
