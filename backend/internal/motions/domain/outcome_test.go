package domain

import "testing"

func TestYrkandeOutcomeMapsRiksdagensWords(t *testing.T) {
	cases := []struct{ kammaren, punkt, want string }{
		{"Avslag", "", OutcomeAvslagen},
		{"avslag", "", OutcomeAvslagen},
		{"Bifall", "", OutcomeBifallen},
		{"Delvis bifall", "", OutcomeDelvis},
		{"Bifall delvis", "", OutcomeDelvis},
		{"Besvarad", "", OutcomeBesvarad},
		{"Besvarad", DecisionAvslag, OutcomeBesvarad}, // kammaren outranks the text
		{"", DecisionAvslag, OutcomeAvslagen},         // empty kammaren falls back to the punkt
		{"", DecisionBifall, OutcomeBifallen},
		{"", DecisionDelvisBifall, OutcomeDelvis},
		{"", DecisionBesvarad, OutcomeBesvarad},
		{"", "", OutcomeEjAvgjord},
		{"Återremitterad", "", "Återremitterad"}, // Riksdagen's own word
		{"  Avslag ", "", OutcomeAvslagen},
	}
	for _, c := range cases {
		if got := YrkandeOutcome(c.kammaren, c.punkt); got != c.want {
			t.Errorf("YrkandeOutcome(%q,%q) = %q, want %q", c.kammaren, c.punkt, got, c.want)
		}
	}
}

func TestMotionOutcomeAggregates(t *testing.T) {
	r := MotionOutcome([]string{OutcomeAvslagen})
	if r.Word != OutcomeAvslagen || r.Counts[OutcomeAvslagen] != 1 {
		t.Errorf("single: %+v", r)
	}
	r = MotionOutcome([]string{OutcomeAvslagen, OutcomeAvslagen})
	if r.Word != OutcomeAvslagen {
		t.Errorf("uniform: %+v", r)
	}
	r = MotionOutcome([]string{OutcomeBifallen, OutcomeAvslagen, OutcomeAvslagen})
	if r.Word != OutcomeDelvis || r.Counts[OutcomeBifallen] != 1 || r.Counts[OutcomeAvslagen] != 2 || r.Total != 3 {
		t.Errorf("mixed: %+v", r)
	}
	r = MotionOutcome([]string{OutcomeAvslagen, OutcomeEjAvgjord})
	if r.Word != OutcomeEjAvgjord {
		t.Errorf("one undecided yrkande keeps the motion undecided: %+v", r)
	}
	r = MotionOutcome(nil)
	if r.Word != OutcomeEjAvgjord {
		t.Errorf("no yrkanden: %+v", r)
	}
	r = MotionOutcome([]string{OutcomeAvslagen, OutcomeBesvarad})
	if r.Word != OutcomeBlandat {
		t.Errorf("avslag+besvarad: %+v", r)
	}
	r = MotionOutcome([]string{"Återremitterad", "Återremitterad"})
	if r.Word != "Återremitterad" {
		t.Errorf("unknown word, uniform: %+v", r)
	}
}

func TestPunktDecisionFromText(t *testing.T) {
	cases := []struct{ text, vinnare, want string }{
		{skU11_1718_p6, "utskottet", DecisionAvslag},
		{skU16_2122_p3, "utskottet", DecisionBifall},
		{skU16_2122_p5, "utskottet", DecisionBifall},
		{"Riksdagen avslår motionerna 2017/18:1 av A (M)", "motförslag", ""},
		{"Riksdagen lägger propositionen till handlingarna", "utskottet", ""},
		{"", "utskottet", ""},
	}
	for _, c := range cases {
		if got := PunktDecisionOf(c.text, c.vinnare); got != c.want {
			t.Errorf("PunktDecisionOf(%.40q,%q) = %q, want %q", c.text, c.vinnare, got, c.want)
		}
	}
}

func TestDecisionMode(t *testing.T) {
	cases := []struct {
		beslutstyp, voteringID string
		want                   DecisionMode
	}{
		{"acklamation", "", ModeAcklamation},
		{"Acklamation", "", ModeAcklamation},
		{"röstning", "0944d710-bc94-493b-981d-9c70815ca406", ModeVotering},
		{"röstning", "", ModeVotering},
		{"", "", ModeEjBeslutad},
		{"bordläggning", "", ModeAnnat},
	}
	for _, c := range cases {
		if got := ModeOf(c.beslutstyp, c.voteringID); got != c.want {
			t.Errorf("ModeOf(%q,%q) = %q, want %q", c.beslutstyp, c.voteringID, got, c.want)
		}
	}
	if ModeAcklamation.Label() != "Avgjord utan omröstning" {
		t.Errorf("label = %q", ModeAcklamation.Label())
	}
}

func TestKnownOutcomeMapsRiksdagensWordsAndFlagsTheRest(t *testing.T) {
	for _, w := range []string{"Avslag", " Bifall", "Delvis bifall", "Besvarad", "Avslag på motionen"} {
		if !KnownOutcome(w) {
			t.Errorf("%q should be mapped", w)
		}
	}
	for _, w := range []string{"Återförvisas till utskottet", "Bordläggs", ""} {
		if KnownOutcome(w) {
			t.Errorf("%q should not be mapped", w)
		}
	}
	if got := MappedOutcome("Återförvisas till utskottet"); got != "Återförvisas till utskottet" {
		t.Errorf("an unknown word must be shown as written, got %q", got)
	}
}
