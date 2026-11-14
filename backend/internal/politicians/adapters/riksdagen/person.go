package riksdagen

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"riksdagskollen/internal/politicians/domain"
)

// rawPerson is one person in a personlista response.
type rawPerson struct {
	IntressentID  string          `json:"intressent_id"`
	Tilltalsnamn  string          `json:"tilltalsnamn"`
	Efternamn     string          `json:"efternamn"`
	Parti         string          `json:"parti"`
	Valkrets      string          `json:"valkrets"`
	BildURL192    string          `json:"bild_url_192"`
	Status        string          `json:"status"`
	Personuppdrag json.RawMessage `json:"personuppdrag"`
	Personuppgift json.RawMessage `json:"personuppgift"`
}

type uppdrag struct {
	OrganKod string          `json:"organ_kod"`
	RollKod  string          `json:"roll_kod"`
	Status   string          `json:"status"`
	Typ      string          `json:"typ"`
	From     string          `json:"from"`
	Tom      string          `json:"tom"`
	Uppgift  json.RawMessage `json:"uppgift"`
}

func (p rawPerson) toPolitician(now time.Time) *domain.Politician {
	uppdragen := decodeList[uppdrag](p.Personuppdrag, "uppdrag")
	pol := &domain.Politician{
		IntressentID: p.IntressentID,
		FirstName:    p.Tilltalsnamn,
		LastName:     p.Efternamn,
		Party:        p.Parti,
		Constituency: p.Valkrets,
		ImageURL:     p.BildURL192,
		IsActive:     strings.Contains(strings.ToLower(p.Status), "tjänstgörande"),
		UpdatedAt:    now,
		Email:        officialEmail(p.Personuppgift),
		Committees:   currentCommittees(uppdragen, now),
	}
	if pol.IsActive {
		pol.MemberSince = memberSince(uppdragen, now)
	}
	return pol
}

// decodeList reads obj[key], which Riksdagen gives as an array, a single
// object, or (when the person has none) omits or replaces obj with "".
func decodeList[T any](raw json.RawMessage, key string) []T {
	if len(raw) == 0 || raw[0] != '{' {
		return nil
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(raw, &obj) != nil {
		return nil
	}
	v := obj[key]
	if len(v) == 0 {
		return nil
	}
	if v[0] == '[' {
		var out []T
		_ = json.Unmarshal(v, &out)
		return out
	}
	var one T
	if json.Unmarshal(v, &one) != nil {
		return nil
	}
	return []T{one}
}

// officialEmail returns the address Riksdagen publishes as "Officiell
// e-postadress". Riksdagen writes "@" as "[på]" to hinder scrapers. Anything
// that does not decode to a plain address is dropped, never guessed.
func officialEmail(raw json.RawMessage) string {
	type uppgift struct {
		Typ     string   `json:"typ"`
		Kod     string   `json:"kod"`
		Uppgift []string `json:"uppgift"`
	}
	for _, u := range decodeList[uppgift](raw, "uppgift") {
		if u.Typ != "eadress" || u.Kod != "Officiell e-postadress" || len(u.Uppgift) == 0 {
			continue
		}
		addr := strings.TrimSpace(strings.Replace(u.Uppgift[0], "[på]", "@", 1))
		if strings.Count(addr, "@") == 1 && !strings.ContainsAny(addr, " \t<>,;") &&
			!strings.HasPrefix(addr, "@") && !strings.HasSuffix(addr, "@") {
			return addr
		}
	}
	return ""
}

// parseTime reads Riksdagen's "2026-09-28 11:00:00" or "2018-09-24". The
// second result is false for an empty or unreadable value.
func parseTime(s string) (time.Time, bool) {
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// isCurrent: started and not ended at now. An empty tom means open-ended.
func (u uppdrag) isCurrent(now time.Time) bool {
	from, ok := parseTime(u.From)
	if !ok || from.After(now) {
		return false
	}
	if u.Tom == "" {
		return true
	}
	tom, ok := parseTime(u.Tom)
	return ok && tom.After(now)
}

// currentCommittees lists the member's current uppdrag in committees (typ
// "uppdrag": utskott and EU-nämnden). One entry per organ and role.
func currentCommittees(us []uppdrag, now time.Time) []domain.Assignment {
	var out []domain.Assignment
	seen := map[string]bool{}
	for _, u := range us {
		if u.Typ != "uppdrag" || u.OrganKod == "" || u.RollKod == "" || !u.isCurrent(now) {
			continue
		}
		key := u.OrganKod + "|" + u.RollKod
		if seen[key] {
			continue
		}
		seen[key] = true
		name := u.OrganKod
		var names []string
		if json.Unmarshal(u.Uppgift, &names) == nil && len(names) > 0 && strings.TrimSpace(names[0]) != "" {
			name = strings.TrimSpace(names[0])
		}
		out = append(out, domain.Assignment{Code: u.OrganKod, Name: name, Role: u.RollKod, From: dateOnly(u.From)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Code != out[j].Code {
			return out[i].Code < out[j].Code
		}
		return out[i].Role < out[j].Role
	})
	return out
}

func dateOnly(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return ""
}

// bridge is the largest gap between two kammaruppdrag that still counts as one
// unbroken run. Riksdagen ends one period at 10:59 and starts the next at
// 11:00, or on the next calendar day.
const bridge = 48 * time.Hour

// memberSince returns the year the current unbroken run as riksdagsledamot
// began, or 0 when the member is not currently one. Time on leave (Ledig) still
// counts as membership; ersättare periods do not. A break longer than bridge
// restarts the run, so a member who left and returned gets the return year.
func memberSince(us []uppdrag, now time.Time) int {
	type span struct{ from, tom time.Time }
	var spans []span
	for _, u := range us {
		if u.Typ != "kammaruppdrag" || u.RollKod != "Riksdagsledamot" ||
			(u.Status != "Tjänstgörande" && u.Status != "Ledig") {
			continue
		}
		from, ok := parseTime(u.From)
		if !ok {
			continue
		}
		tom := now.Add(100 * 365 * 24 * time.Hour)
		if u.Tom != "" {
			t, ok := parseTime(u.Tom)
			if !ok {
				continue
			}
			tom = t
		}
		spans = append(spans, span{from, tom})
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].from.After(spans[j].from) })
	start := time.Time{}
	reach := time.Time{} // earliest start of the run so far
	for _, s := range spans {
		switch {
		case reach.IsZero():
			if s.from.After(now) || !s.tom.After(now) {
				continue // future or ended: the run must contain now
			}
			reach = s.from
		case s.tom.Add(bridge).Before(reach):
			// gap: the run found so far is the current one
			return start.Year()
		default:
			if s.from.Before(reach) {
				reach = s.from
			}
		}
		start = reach
	}
	if reach.IsZero() {
		return 0
	}
	return start.Year()
}
