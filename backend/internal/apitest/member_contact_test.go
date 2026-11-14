package apitest

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Recorded 2026-11-14 from data.riksdagen.se/personlista (see
// fixtures/member-contact/personlista.json).
const (
	malinLarssonID   = "0203877548421" // S, MJU ledamot, riksdagsledamot sedan 2018
	mikaelDambergID  = "014744660015"  // S, FiU ordförande, ledamot sedan 2002 (tjänstledig 2014-2022)
	memberContactDay = "2026-11-14"
)

type memberPage struct {
	Email        string `json:"email"`
	RiksdagenURL string `json:"riksdagenUrl"`
	MemberSince  int    `json:"memberSince"`
	Committees   []struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Role string `json:"role"`
	} `json:"committees"`
}

func startMemberContact(t *testing.T) *Harness {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "fixtures", "member-contact")
	day, _ := time.Parse("2006-01-02", memberContactDay)
	h := Start(t, dir, WithNow(day.Add(12*time.Hour)))
	h.RunWorker(context.Background(), "fetch-politicians")
	return h
}

func TestMemberPageLinksToRiksdagenAndShowsOfficialEmail(t *testing.T) {
	h := startMemberContact(t)
	var p memberPage
	getJSON(t, h, "/api/politicians/"+malinLarssonID, &p)
	if p.RiksdagenURL != "https://www.riksdagen.se/sv/ledamoter-och-partier/ledamot/malin-larsson_"+malinLarssonID+"/" {
		t.Fatalf("riksdagenUrl = %q", p.RiksdagenURL)
	}
	if p.Email != "malin.larsson@riksdagen.se" {
		t.Fatalf("email = %q", p.Email)
	}
}

func TestMemberPageShowsCurrentCommitteesOnly(t *testing.T) {
	h := startMemberContact(t)
	var p memberPage
	getJSON(t, h, "/api/politicians/"+malinLarssonID, &p)
	if len(p.Committees) != 1 || p.Committees[0].Code != "MJU" ||
		p.Committees[0].Role != "Ledamot" || p.Committees[0].Name != "Miljö- och jordbruksutskottet" {
		t.Fatalf("malin committees = %+v", p.Committees)
	}
	var d memberPage
	getJSON(t, h, "/api/politicians/"+mikaelDambergID, &d)
	if len(d.Committees) != 1 || d.Committees[0].Code != "FiU" || d.Committees[0].Role != "Ordförande" {
		t.Fatalf("damberg committees = %+v", d.Committees)
	}
}

func TestMemberSinceFollowsUnbrokenMandateChain(t *testing.T) {
	h := startMemberContact(t)
	var p, d memberPage
	getJSON(t, h, "/api/politicians/"+malinLarssonID, &p)
	getJSON(t, h, "/api/politicians/"+mikaelDambergID, &d)
	if p.MemberSince != 2018 {
		t.Fatalf("malin memberSince = %d", p.MemberSince)
	}
	// Leave of absence (status Ledig) does not end membership.
	if d.MemberSince != 2002 {
		t.Fatalf("damberg memberSince = %d", d.MemberSince)
	}
}

func TestMemberContactSyncIsIdempotent(t *testing.T) {
	h := startMemberContact(t)
	h.RunWorker(context.Background(), "fetch-politicians")
	var p memberPage
	getJSON(t, h, "/api/politicians/"+malinLarssonID, &p)
	if len(p.Committees) != 1 || p.MemberSince != 2018 {
		t.Fatalf("after second sync: %+v", p)
	}
}
