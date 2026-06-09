package main

import (
	"context"
	"fmt"
	"strconv"

	politicians "riksdagskollen/internal/politicians"
	polrd "riksdagskollen/internal/politicians/adapters/riksdagen"
	poldomain "riksdagskollen/internal/politicians/domain"
	polports "riksdagskollen/internal/politicians/ports"
	speechrd "riksdagskollen/internal/speeches/adapters/riksdagen"
	speechports "riksdagskollen/internal/speeches/ports"
	votes "riksdagskollen/internal/votes"
	voterd "riksdagskollen/internal/votes/adapters/riksdagen"
	votedomain "riksdagskollen/internal/votes/domain"
	voteports "riksdagskollen/internal/votes/ports"
)

// riksdagParties are the eight Riksdag parties whose seat counts we verify.
var riksdagParties = []string{"S", "M", "SD", "C", "V", "KD", "L", "MP"}

// sampleVotes are fixed (beteckning, punkt) pairs whose tallies we spot-check.
// AU10 p1 is a near-full-chamber vote present in the seeded DB (340 rows);
// FiU1 from the original sketch is not ingested locally. Note the upstream
// side fetches by beteckning (all punkter), so multi-punkt betänkanden show a
// scope difference vs our single-punkt tally — expected, not a data bug.
var sampleVotes = []struct{ Beteckning, Punkt string }{
	{"AU10", "1"},
}

// memberStatusSitting is the FetchMembers status value for currently-sitting
// members. It maps to riksdagen's `rdlstatus` query param; the politicians
// service uses exactly this value (see internal/politicians/service.go).
const memberStatusSitting = "tjanstgorande"

// speechSession is the current Riksdag session used to fetch a small recent
// speech batch, mirroring the speeches ingestion worker.
const speechSession = "2024/25"

func riksdagenChecks(
	ctx context.Context,
	polSvc *politicians.Service, polCl *polrd.Client,
	voteSvc *votes.Service, voteCl *voterd.Client,
	speechCl *speechrd.Client,
) []Result {
	var out []Result

	// Party seat totals (per party).
	for _, party := range riksdagParties {
		ours, oerr := polSvc.List(ctx, polports.ListFilter{Party: party, ActiveOnly: true, Page: 1, PageSize: 1})
		upstream, uerr := polCl.FetchMembers(ctx, party, memberStatusSitting)
		r := Result{Tier: "riksdag", Entity: "Party " + party, Metric: "seat total", SourceID: "riksdagen"}
		switch {
		case oerr != nil:
			r.Status, r.Upstream = StatusError, oerr.Error()
		case uerr != nil:
			r.Status, r.Ours, r.Upstream = StatusError, strconv.Itoa(ours.Total), uerr.Error()
		default:
			r.Ours = strconv.Itoa(ours.Total)
			// rdlstatus does not filter server-side, so the response includes
			// former members; count only those the client marks IsActive.
			r.Upstream = strconv.Itoa(countActive(upstream))
			r.Status = classify(r.Ours, r.Upstream, false)
		}
		out = append(out, r)
	}

	// Roster grand total.
	allOurs, oerr := polSvc.List(ctx, polports.ListFilter{ActiveOnly: true, Page: 1, PageSize: 1})
	r := Result{Tier: "riksdag", Entity: "Riksdag", Metric: "active member roster", SourceID: "riksdagen"}
	if oerr != nil {
		r.Status, r.Upstream = StatusError, oerr.Error()
	} else {
		all, uerr := polCl.FetchMembers(ctx, "", memberStatusSitting)
		r.Ours = strconv.Itoa(allOurs.Total)
		if uerr != nil {
			r.Status, r.Upstream = StatusError, uerr.Error()
		} else {
			r.Upstream = strconv.Itoa(countActive(all))
			r.Status = classify(r.Ours, r.Upstream, false)
		}
	}
	out = append(out, r)

	// Sample vote outcomes.
	for _, sv := range sampleVotes {
		oursVotes, oerr := voteSvc.ListByBeteckning(ctx, sv.Beteckning, sv.Punkt)
		upVotes, uerr := voteCl.FetchVotes(ctx, voteports.FetchVotesFilter{Beteckning: sv.Beteckning})
		entity := fmt.Sprintf("Vote %s p%s", sv.Beteckning, sv.Punkt)
		if oerr != nil || uerr != nil {
			msg := ""
			if oerr != nil {
				msg = oerr.Error()
			} else {
				msg = uerr.Error()
			}
			out = append(out, Result{Tier: "riksdag", Entity: entity, Metric: "vote tally", SourceID: "riksdagen", Status: StatusError, Upstream: msg})
			continue
		}
		// Informational: our side is a single förslagspunkt, while the upstream
		// FetchVotes returns the whole betänkande and does not reliably expose a
		// matching punkt field, so the two tallies are not directly comparable.
		// Recorded for eyeballing, not flagged for review.
		ourTally := tally(oursVotes)
		upTally := tally(upVotes)
		out = append(out, Result{
			Tier: "riksdag", Entity: entity, Metric: "vote tally Ja/Nej/Avst/Frånv (ours=1 punkt, upstream=hela betänkandet)", SourceID: "riksdagen",
			Ours: ourTally, Upstream: upTally, Status: classify(ourTally, upTally, true),
		})
	}

	// Speech counts (informational/volatile).
	recent, err := speechCl.FetchSpeeches(ctx, recentSpeechFilter())
	sc := Result{Tier: "riksdag", Entity: "Riksdag", Metric: "anföranden (senaste hämtning)", SourceID: "riksdagen", Ours: "n/a (volatile)"}
	if err != nil {
		sc.Status, sc.Upstream = StatusError, err.Error()
	} else {
		sc.Upstream = strconv.Itoa(len(recent))
		sc.Status = classify(sc.Ours, sc.Upstream, true) // always Info
	}
	out = append(out, sc)

	return out
}

// recentSpeechFilter builds the smallest FetchSpeechesFilter that returns a
// recent batch — the current session with a small page size, mirroring the
// speeches ingestion worker (internal/ingestion/workers/speeches.go).
func recentSpeechFilter() speechports.FetchSpeechesFilter {
	return speechports.FetchSpeechesFilter{
		Session: speechSession,
		Size:    20,
	}
}

// countActive returns how many members the riksdagen client marked as
// currently serving (IsActive). The rdlstatus query param does not filter
// server-side, so the raw list also contains former members.
func countActive(members []*poldomain.Politician) int {
	n := 0
	for _, m := range members {
		if m.IsActive {
			n++
		}
	}
	return n
}

// tally counts votes into a canonical "Ja/Nej/Avstår/Frånvarande" string.
func tally(vs []*votedomain.Vote) string {
	var ja, nej, av, fr int
	for _, v := range vs {
		switch v.VoteResult {
		case votedomain.VoteJa:
			ja++
		case votedomain.VoteNej:
			nej++
		case votedomain.VoteAvstar:
			av++
		case votedomain.VoteFranvarande:
			fr++
		}
	}
	return fmt.Sprintf("%d/%d/%d/%d", ja, nej, av, fr)
}
