package apitest

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const malmerStenergardID = "0572788110724" // Maria Malmer Stenergard (M)

func TestMemberMotionsIncludeMotion3293WithOutcomeAvslagen(t *testing.T) {
	h := startChain(t, chainDir(t, "decision-chain"), "{2017/18,2018/19,2021/22}")
	runChain(h)
	var b struct {
		Data []struct {
			Riksmote   string `json:"riksmote"`
			Beteckning string `json:"beteckning"`
			Outcome    struct {
				Word string `json:"word"`
			} `json:"outcome"`
		} `json:"data"`
	}
	getJSON(t, h, "/api/motions?author="+malmerStenergardID, &b)
	for _, m := range b.Data {
		if m.Riksmote == "2017/18" && m.Beteckning == "3293" {
			if m.Outcome.Word != "Avslagen" {
				t.Fatalf("outcome = %q", m.Outcome.Word)
			}
			return
		}
	}
	t.Fatalf("2017/18:3293 not in %+v", b.Data)
}

func TestSvantessonPageListsFraga339AsAnswered(t *testing.T) {
	h := startQuestions(t, chainDir(t, "questions"))
	q := find(memberQuestions(t, h, svantessonID), "fraga", "339")
	if q == nil || q.Role != "answerer" || q.AnswerDate == nil {
		t.Fatalf("fråga 339 not answered on Svantesson's page: %+v", q)
	}
}

func TestMemberBallotsMarkAvvikareWithSharedDefinition(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "fixtures", "one-votering")
	h := Start(t, dir, WithNow(time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	h.RunWorker(ctx, "fetch-politicians")
	h.RunWorker(ctx, "fetch-votes")
	type page struct {
		Data []struct {
			Beteckning    string `json:"beteckning"`
			Forslagspunkt string `json:"forslagspunkt"`
			Session       string `json:"session"`
			Avvikare      bool   `json:"avvikare"`
			Partilinje    string `json:"partilinje"`
		} `json:"data"`
	}
	var p page
	getJSON(t, h, "/api/politicians/"+recordedPoliticianID+"/votes", &p)
	if len(p.Data) == 0 || p.Data[0].Avvikare || p.Data[0].Partilinje != "Nej" {
		t.Fatalf("before flip: %+v", p.Data)
	}
	if _, err := h.DB.Exec(ctx,
		`UPDATE ballots SET vote_result = 'Ja' WHERE politician_id = $1`, recordedPoliticianID); err != nil {
		t.Fatal(err)
	}
	p = page{}
	getJSON(t, h, "/api/politicians/"+recordedPoliticianID+"/votes", &p)
	if !p.Data[0].Avvikare || p.Data[0].Session == "" || p.Data[0].Forslagspunkt == "" {
		t.Fatalf("after flip: %+v", p.Data[0])
	}
}
