package domain

import (
	"reflect"
	"testing"
)

// Förslag text below is copied from data.riksdagen.se/dokumentstatus
// (dokutskottsforslag), recorded 2026-11-10, with the HTML breaks and entities
// as served. "H501SkU11" = 2017/18:SkU11, "H901SkU16" = 2021/22:SkU16,
// "H601SkU11" = 2018/19:SkU11, "H701SkU21" = 2019/20:SkU21.
const (
	skU11_1718_p6 = "<BR/>Riksdagen avslår motionerna \n\n2017/18:1380 av Suzanne Svensson och Peter Jeppsson (båda S), \n\n2017/18:3002 av Jonas Jacobsson Gjörtler (M), \n\n2017/18:3293 av Maria Malmer Stenergard m.fl. (M) och \n\n2017/18:3748 av Penilla Gunther m.fl. (KD) yrkande 18.<BR/><BR/>"
	skU16_2122_p6 = "<BR/>Riksdagen avslår motionerna \n\n2021/22:3312 av Jan Ericson (M), \n\n2021/22:3637 av Niklas Wykman m.fl. (M) yrkande 6, \n\n2021/22:3638 av Niklas Wykman m.fl. (M) yrkande 9, \n\n2021/22:3685 av Per Åsling m.fl. (C) yrkande 17 och \n\n2021/22:4217 av Camilla Brodin m.fl. (KD) yrkande 4.<BR/><BR/>"
	skU16_2122_p5 = "Riksdagen ställer sig bakom det som utskottet anför om att skyndsamt se över skattelagstiftningen i syfte att säkerställa att den inte i onödan minskar incitamenten att omvandla lokaler och kontor till bostäder och tillkännager detta för regeringen.<BR/>Därmed bifaller riksdagen motion \n\n2021/22:3708 av Carl-Oskar Bohlin m.fl. (M) yrkande 15 och \n\nbifaller delvis motion \n\n2021/22:4222 av Larry Söder m.fl. (KD) yrkande 4 i denna del.<BR/><BR/>"
	skU16_2122_p3 = "Riksdagen ställer sig bakom det som utskottet anför om att regeringen ska utreda hur frågan om intern moms mellan föreningar i samma organisation kan lösas utan att orsaka snedvridning av konkurrensen i strid med EU-rätten och utan att öka risken för fusk med momssystemet, och tillkännager detta för regeringen.\nDärmed bifaller riksdagen motionerna \n\n2021/22:2288 av Malin Danielsson (L) och \n\n2021/22:2593 av Vasiliki Tsouplaki m.fl. (V) yrkande 5.<BR/><BR/>"
	skU16_2122_p1 = "<BR/>Riksdagen avslår motionerna \n\n2021/22:3938 av Magnus Jacobsson m.fl. (KD), \n\n2021/22:3954 av Eric Westroth m.fl. (SD) yrkandena 8 och 11, \n\n2021/22:4217 av Camilla Brodin m.fl. (KD) yrkande 2.<BR/><BR/>"
	skU21_1920_p2 = "<BR/>Riksdagen avslår motionerna \n\n2019/20:1740 av Per Åsling (C), \n\n2019/20:2551 av Per Åsling (C) yrkande 5, \n\n2019/20:2760 av Magnus Oscarsson m.fl. (KD) yrkandena 36 och 63-66, \n\n2019/20:3241 av Hampus Hagman m.fl. (KD) yrkande 9, \n\n2019/20:3247 av Kristina Yngwe m.fl. (C) yrkandena 6 och 7 samt \n\n2019/20:3308 av Rebecka Le Moine (MP) yrkande 2.<BR/><BR/>"
	skU21_1920_p7 = "<BR/>Riksdagen avslår motionerna \n\n2019/20:214 av Hampus Hagman (KD), \n\n2019/20:608 av Roger Hedlund m.fl. (SD) yrkandena 19-23, \n\n2019/20:1002 av Hanna Westerén (S), \n\n2019/20:2149 av Momodou Malcolm Jallow m.fl. (V) yrkande 8.<BR/><BR/>"
	skU11_1819_p4 = "<BR/>Riksdagen avslår motionerna \n\n2018/19:2490 av Per Åsling (C), \n\n2018/19:2736 av Magnus Oscarsson m.fl. (KD) yrkandena 33 och 56-59 samt \n\n2018/19:2761 av Hampus Hagman m.fl. (KD) yrkande 8.<BR/><BR/>"
	skU11_1819_p1 = "<BR/>Riksdagen avslår motionerna \n\n2018/19:2899 av Maria Malmer Stenergard m.fl. (M) yrkande 6, \n\n2018/19:2900 av Jessica Rosencrantz m.fl. (M) yrkande 26, \n\n2018/19:2902 av Jessica Rosencrantz m.fl. (M) yrkande 6 och \n\n2018/19:2914 av Larry Söder m.fl. (KD) yrkande 4.<BR/><BR/>"
)

func link(rm, bet string, nr int, punktTyp string, punkter ...Punkt) []PunktLink {
	return LinkYrkande(YrkandeRef{Riksmote: rm, Beteckning: bet, Number: nr, BehandlasIPunkt: punktTyp}, punkter)
}

func TestLinkerWholeMotionCitedWithoutYrkande(t *testing.T) {
	// 2017/18:3293 is cited bare, so every yrkande of it was decided in p6.
	got := link("2017/18", "3293", 1, "", Punkt{Number: "5", Text: "irrelevant"}, Punkt{Number: "6", Text: skU11_1718_p6})
	want := []PunktLink{{Punkt: "6", How: LinkedByForslagText, Decision: DecisionAvslag}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLinkerSingleYrkandeCitation(t *testing.T) {
	p := Punkt{Number: "6", Text: skU16_2122_p6}
	if got := link("2021/22", "3638", 9, "", p); len(got) != 1 || got[0].Punkt != "6" {
		t.Fatalf("yrkande 9 = %+v", got)
	}
	// Yrkande 1 of 3638 is not among those cited ("yrkande 9" only).
	if got := link("2021/22", "3638", 1, "", p); len(got) != 0 {
		t.Fatalf("yrkande 1 must not link, got %+v", got)
	}
	// 2017/18:3748 yrkande 18 only; yrkande 1 is not decided here.
	if got := link("2017/18", "3748", 18, "", Punkt{Number: "6", Text: skU11_1718_p6}); len(got) != 1 {
		t.Fatalf("3748 y18 = %+v", got)
	}
	if got := link("2017/18", "3748", 17, "", Punkt{Number: "6", Text: skU11_1718_p6}); len(got) != 0 {
		t.Fatalf("3748 y17 must not link, got %+v", got)
	}
}

func TestLinkerListForm(t *testing.T) {
	p := Punkt{Number: "1", Text: skU16_2122_p1}
	for nr, want := range map[int]bool{8: true, 11: true, 9: false, 1: false} {
		got := link("2021/22", "3954", nr, "", p)
		if (len(got) == 1) != want {
			t.Errorf("3954 yrkande %d: got %+v, want linked=%v", nr, got, want)
		}
	}
}

func TestLinkerRangeForms(t *testing.T) {
	cases := []struct {
		rm, bet string
		text    string
		in, out []int
	}{
		{"2019/20", "2760", skU21_1920_p2, []int{36, 63, 64, 65, 66}, []int{37, 62, 67}},
		{"2019/20", "608", skU21_1920_p7, []int{19, 20, 23}, []int{18, 24}},
		{"2018/19", "2736", skU11_1819_p4, []int{33, 56, 59}, []int{34, 55, 60}},
		{"2019/20", "3247", skU21_1920_p2, []int{6, 7}, []int{5, 8}},
	}
	for _, c := range cases {
		for _, n := range c.in {
			if got := link(c.rm, c.bet, n, "", Punkt{Number: "2", Text: c.text}); len(got) != 1 {
				t.Errorf("%s:%s yrkande %d should link, got %+v", c.rm, c.bet, n, got)
			}
		}
		for _, n := range c.out {
			if got := link(c.rm, c.bet, n, "", Punkt{Number: "2", Text: c.text}); len(got) != 0 {
				t.Errorf("%s:%s yrkande %d should not link, got %+v", c.rm, c.bet, n, got)
			}
		}
	}
}

func TestLinkerDelvisAndIDennaDel(t *testing.T) {
	p := Punkt{Number: "5", Text: skU16_2122_p5}
	got := link("2021/22", "4222", 4, "", p)
	if len(got) != 1 || got[0].Decision != DecisionDelvisBifall {
		t.Fatalf("bifaller delvis: got %+v", got)
	}
	// The motion before it in the same sentence is a plain bifall.
	got = link("2021/22", "3708", 15, "", p)
	if len(got) != 1 || got[0].Decision != DecisionBifall {
		t.Fatalf("bifaller: got %+v", got)
	}
	// A bifall sentence ahead of the citation list still applies to its motions.
	got = link("2021/22", "2593", 5, "", Punkt{Number: "3", Text: skU16_2122_p3})
	if len(got) != 1 || got[0].Decision != DecisionBifall {
		t.Fatalf("p3: got %+v", got)
	}
}

func TestLinkerYrkandeDelvisAfterNumber(t *testing.T) {
	// The "yrkande 2 delvis" form. No recorded betänkande in the fixtures
	// carries it; this is the form the issue names, built on a recorded frame.
	text := "Riksdagen bifaller motion \n\n2021/22:100 av A B (M) yrkande 2 delvis och \n\n2021/22:101 av C D (S) yrkande 1."
	got := link("2021/22", "100", 2, "", Punkt{Number: "1", Text: text})
	if len(got) != 1 || got[0].Decision != DecisionDelvisBifall {
		t.Fatalf("got %+v", got)
	}
	got = link("2021/22", "101", 1, "", Punkt{Number: "1", Text: text})
	if len(got) != 1 || got[0].Decision != DecisionBifall {
		t.Fatalf("neighbour must not inherit delvis: %+v", got)
	}
}

func TestLinkerUsesBehandlasIPunktWhenPresent(t *testing.T) {
	// The explicit punkt wins; the text is not consulted.
	got := link("2017/18", "3293", 1, "6", Punkt{Number: "6", Text: "ingen hänvisning här"})
	want := []PunktLink{{Punkt: "6", How: LinkedByBehandlasIPunkt}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	// A punkt that the betänkande does not have falls back to the text.
	got = link("2017/18", "3293", 1, "99", Punkt{Number: "6", Text: skU11_1718_p6})
	if len(got) != 1 || got[0].Punkt != "6" || got[0].How != LinkedByForslagText {
		t.Fatalf("fallback: %+v", got)
	}
}

func TestLinkerEmptyBehandlasIPunktParsesText2018Yrkande26(t *testing.T) {
	// 2018/19:2900 yrkande 26 has an empty behandlas_i_punkt.
	got := link("2018/19", "2900", 26, "", Punkt{Number: "1", Text: skU11_1819_p1}, Punkt{Number: "4", Text: skU11_1819_p4})
	if len(got) != 1 || got[0].Punkt != "1" || got[0].How != LinkedByForslagText || got[0].Decision != DecisionAvslag {
		t.Fatalf("got %+v", got)
	}
}

func TestLinkerMissingFromEveryPunktIsNoLink(t *testing.T) {
	got := link("2017/18", "4202", 1, "", Punkt{Number: "6", Text: skU11_1718_p6}, Punkt{Number: "1", Text: skU16_2122_p1})
	if len(got) != 0 {
		t.Fatalf("must not guess, got %+v", got)
	}
}

func TestLinkerDoesNotConfuseBeteckningPrefixes(t *testing.T) {
	// 2017/18:380 must not match 2017/18:3809 or 1380.
	p := Punkt{Number: "6", Text: skU11_1718_p6}
	if got := link("2017/18", "380", 1, "", p); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
	if got := link("2017/18", "138", 1, "", p); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}
