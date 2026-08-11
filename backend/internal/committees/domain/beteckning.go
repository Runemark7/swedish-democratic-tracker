package domain

import "strings"

// canonicalCodes maps an upper-cased committee code to its canonical spelling.
//
// 2002-2006 beteckningar are ALL-CAPS ("UBU14", "JUU9"). Without this, the same
// committee appears twice in a derived list — once per casing — and ~1 070
// voteringar hide under codes no lookup recognises.
var canonicalCodes = map[string]string{
	"AU": "AU", "CU": "CU", "FIU": "FiU", "FÖU": "FöU", "JUU": "JuU",
	"KRU": "KrU", "KU": "KU", "MJU": "MJU", "NU": "NU", "SFU": "SfU",
	"SKU": "SkU", "SOU": "SoU", "TU": "TU", "UBU": "UbU", "UU": "UU",
	// Joint committee: Utrikes- och försvarsutskottet.
	"UFÖU": "UFöU",
	// Abolished in 2006, present in the 2002-2006 record.
	"LU": "LU", "BOU": "BoU",
}

// CommitteeCode extracts the canonical committee code from a beteckning.
//
// A beteckning runs code-then-number ("SoU12", "AU1y"). The code is everything
// up to the first digit. Returns "" when there is no code — a votering on a
// motion rather than a betänkande carries none, and that absence is reported
// rather than guessed at.
func CommitteeCode(beteckning string) string {
	end := -1
	for i, r := range beteckning {
		if r >= '0' && r <= '9' {
			end = i
			break
		}
	}
	raw := beteckning
	if end >= 0 {
		raw = beteckning[:end]
	}
	return Canonical(raw)
}

// Canonical returns the canonical spelling of a bare committee code.
//
// Every committee code ends in "U" (utskott), which is what distinguishes a code
// from any other beteckning prefix: "prop." must not be mistaken for one.
func Canonical(rawCode string) string {
	upper := strings.ToUpper(rawCode)
	if rawCode == "" || !strings.HasSuffix(upper, "U") {
		return ""
	}
	if canon, ok := canonicalCodes[upper]; ok {
		return canon
	}
	// An unknown code is returned as-is rather than dropped. A code we have
	// never seen is a fact about the record, and hiding it would silently
	// remove real voteringar from every derived list.
	return rawCode
}
