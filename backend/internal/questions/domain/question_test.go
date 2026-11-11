package domain

import (
	"reflect"
	"testing"
)

func TestCitedMotions(t *testing.T) {
	cases := []struct {
		name, text string
		want       []string
	}{
		{"motion beteckning", "Riksdagen avslog motion 2017/18:3293 i mars.", []string{"2017/18:3293"}},
		{"kommittémotion", "I kommittémotionen 2021/22:3638 yrkas", []string{"2021/22:3638"}},
		{"list", "motionerna 2021/22:1 och 2021/22:2", []string{"2021/22:1", "2021/22:2"}},
		{"mot. abbreviation", "(mot. 2018/19:2900)", []string{"2018/19:2900"}},
		{"prose only", "i en kommittémotion från Moderaterna redan 2017", nil},
		{"proposition is not a motion", "regeringens proposition 2022/23:100 om skatt", nil},
		{"motion then proposition", "motionen i prop. 2022/23:100", nil},
		{"own beteckning", "Fråga 2022/23:339 och motion 2022/23:339", nil},
		{"bare beteckning", "se 2018/19:2900", nil},
		{"far from keyword", "motion som lades fram för att reformera skattesystemet på ett sätt som tar hänsyn 2018/19:2900", nil},
		{"dedup", "motion 2018/19:5 samt motion 2018/19:5", []string{"2018/19:5"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CitedMotions(c.text, "2022/23", "339")
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}

func TestResolveKeepsOnlyHeldMotionsWithTheirRule(t *testing.T) {
	q := Question{DokID: "Q1", Riksmote: "2022/23", Beteckning: "7",
		BodyText:   "motion 2021/22:3638 och motion 2021/22:9999",
		MotionRefs: []string{"2017/18:3293"}}
	held := map[string]string{"2017/18:3293": "M1", "2021/22:3638": "M2"}
	got := Resolve(q, func(r string) (string, bool) { id, ok := held[r]; return id, ok })
	want := []Link{{"Q1", "M1", RuleReference}, {"Q1", "M2", RuleCited}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
