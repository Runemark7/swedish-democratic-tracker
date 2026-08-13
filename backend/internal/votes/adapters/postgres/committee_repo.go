package postgres

import (
	"context"

	committeedomain "riksdagskollen/internal/committees/domain"
	"riksdagskollen/internal/votes/ports"
)

// The unit of this query is votering_id, which is the whole key: beteckning
// numbering restarts each riksmöte, so keying on beteckning+forslagspunkt
// silently folds AU5/2022-23 into AU5/2024-25 — both undercounting the record
// and blending two decisions' party positions into one row.
//
// Scoped to a mandate period via mandate_periods, the same table and join
// pattern committees/adapters/postgres/repository.go's countVoteringar uses:
// JOIN mandate_periods mp ON v.session = ANY(mp.riksmoten) WHERE mp.code = $1.
// Without it this spans the whole votes table, which today holds only the
// 2022-2026 period and so happens to agree with a scoped query — but the
// record already anticipates 2002-2006 (canonicalCodes carries LU/BoU marked
// as present in it), and the two periods must never be blended.
//
// starts_with(upper(beteckning), upper($2)) is only a narrowing predicate,
// never the authoritative filter. A raw prefix like "U" is a prefix of UU,
// UbU and UFöU alike, and starts_with alone would serve all three as one
// committee's record. The authoritative code is always derived in Go by
// committeedomain.CommitteeCode and compared against the requested code in
// the fold below — exactly what committees/adapters/postgres/repository.go's
// countVoteringar already does — so a beteckning that merely starts with the
// same letters can never be counted under the wrong committee.
//
// party <> '-' excludes ballots cast under no party affiliation (vacant or
// independent seats — 94 of AU's 1 022 position rows, 7 740 record-wide):
// RÖSTAT is a per-party list and "-" is not a party. Those votes stay in the
// record and on the decision page; they just don't occupy a party's slot
// here.
//
// Ties are surfaced, not resolved. counts/maxn/pos together compute, per
// (votering, party), which vote_result(s) hold the highest count: exactly one
// survives the join to maxn and it is that party's position; two or more
// survive and the party is reported "Delad" rather than resolved by whichever
// vote_result sorts first alphabetically. That silent tie-break used to
// publish "Frånvarande" for a party that was tied 8 Frånvarande / 8 Ja with
// Ja — a flatly wrong statement about how half the party voted.
//
// Ordering: votes carry no decision date (only system_datum, which is when
// Riksdagen last touched the row), so chronological order is unavailable and
// must not be faked. Order by riksmöte, then the numeric part of the beteckning
// — regexp_replace strips non-digits so "AU10" → 10 sorts after "AU9" → 9,
// which ordering the raw text would get backwards. NULLIF guards the cast: a
// beteckning with no digits would otherwise make an empty string raise on
// the ::int cast. votering_id is the final tie-break: 18 förslagspunkter in
// 2022-2026 were genuinely decided by two separate voteringar, and without a
// deterministic last key their relative order would be unspecified across
// paged requests.
const committeeVoteringQuery = `
WITH counts AS (
	SELECT v.votering_id, v.party, v.vote_result, count(*) AS n
	FROM votes v
	JOIN mandate_periods mp ON v.session = ANY(mp.riksmoten)
	WHERE mp.code = $1
	  AND starts_with(upper(v.beteckning), upper($2))
	  AND v.party <> '-'
	GROUP BY v.votering_id, v.party, v.vote_result
),
maxn AS (
	SELECT votering_id, party, max(n) AS maxn
	FROM counts
	GROUP BY votering_id, party
),
pos AS (
	SELECT c.votering_id, c.party,
	       CASE WHEN count(*) > 1 THEN 'Delad'
	            ELSE min(c.vote_result)
	       END AS position
	FROM counts c
	JOIN maxn m ON m.votering_id = c.votering_id AND m.party = c.party AND c.n = m.maxn
	GROUP BY c.votering_id, c.party
),
pts AS (
	SELECT v.votering_id,
	       min(v.beteckning)                   AS beteckning,
	       min(v.forslagspunkt)                AS forslagspunkt,
	       min(v.session)                      AS session,
	       COALESCE(max(v.document_title), '') AS document_title
	FROM votes v
	JOIN mandate_periods mp ON v.session = ANY(mp.riksmoten)
	WHERE mp.code = $1
	  AND starts_with(upper(v.beteckning), upper($2))
	GROUP BY v.votering_id
)
SELECT pt.votering_id, pt.beteckning, pt.forslagspunkt, pt.document_title, pt.session,
       pos.party, pos.position
FROM pts pt
JOIN pos ON pos.votering_id = pt.votering_id
ORDER BY pt.session,
         NULLIF(regexp_replace(pt.beteckning, '\D', '', 'g'), '')::int,
         NULLIF(regexp_replace(pt.forslagspunkt, '\D', '', 'g'), '')::int,
         pos.party,
         pt.votering_id`

// PeriodExists reports whether the mandate period is one the record knows.
//
// Identical query to committees/adapters/postgres/repository.go's
// PeriodExists: an unrecognised period (e.g. "2018-2022", before this record
// begins) must be refused rather than answered with an empty page, which
// reads as "this committee decided nothing" — a claim about the record this
// endpoint is in no position to make.
func (r *Repository) PeriodExists(ctx context.Context, periodCode string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM mandate_periods WHERE code = $1)`, periodCode,
	).Scan(&exists)
	return exists, err
}

// ListByCommitteeWithPositions returns the committee's voteringar in one
// mandate period, each with the dominant position per party.
//
// periodCode is required and applied exactly like /committees/{code}: an
// unscoped query would blend mandate periods together the moment the record
// holds more than one.
func (r *Repository) ListByCommitteeWithPositions(ctx context.Context, periodCode, code string, limit, offset int) ([]ports.CommitteeVotering, int, error) {
	rows, err := r.db.Query(ctx, committeeVoteringQuery, periodCode, code)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Rows arrive one per (votering, party); fold them into voteringar. The fold
	// key is votering_id for the same reason the SQL groups on it — folding on
	// beteckning+forslagspunkt would merge riksmöten back together here even if
	// the query kept them apart.
	var out []ports.CommitteeVotering
	idx := map[string]int{}
	for rows.Next() {
		var vid, bet, punkt, title, rm, party, result string
		if err := rows.Scan(&vid, &bet, &punkt, &title, &rm, &party, &result); err != nil {
			return nil, 0, err
		}

		// Authoritative committee filter. starts_with($2) only narrowed the SQL
		// scan; committeedomain.CommitteeCode(bet) is the record's own rule for
		// which committee a beteckning belongs to, and a row whose derived code
		// disagrees with the requested code is dropped here rather than trusted
		// because it happened to share a prefix.
		if committeedomain.CommitteeCode(bet) != code {
			continue
		}

		i, ok := idx[vid]
		if !ok {
			out = append(out, ports.CommitteeVotering{
				VoteringID:    vid,
				Beteckning:    bet,
				Forslagspunkt: punkt,
				DocumentTitle: title,
				Riksmote:      rm,
			})
			i = len(out) - 1
			idx[vid] = i
		}
		out[i].PartyPositions = append(out[i].PartyPositions,
			ports.PartyPosition{Party: party, Position: result})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// The total is the folded length, not a second COUNT query. Two queries with
	// two copies of the predicate can drift apart, and a total that disagrees
	// with the rows beneath it is exactly the kind of authoritative-looking wrong
	// number this site exists to avoid. Filtering in Go above (rather than after
	// this point) means that guarantee still holds: total is never counted
	// before the authoritative committee filter runs.
	total := len(out)

	// Mark the förslagspunkter the record decided with more than one votering,
	// over the whole slice and BEFORE paging. Computed per page instead, the
	// mark is wrong exactly where it matters most: the two voteringar for NU7
	// punkt 2 (2025/26) sit at indices 149 and 150, so with PAGE_SIZE 50 each
	// page holds one of the pair, sees no duplicate, and marks neither. The
	// reader then gets two rows with the same beteckning, punkt, title and
	// riksmöte but opposite party positions, and no reason for the difference.
	// This function already materialises every votering, so it is the only
	// place that can see the pair at all.
	perPoint := map[string]int{}
	for i := range out {
		perPoint[forslagspunktKey(out[i])]++
	}
	for i := range out {
		out[i].DecidedByMultipleVoteringar = perPoint[forslagspunktKey(out[i])] > 1
	}

	// Page after folding, so a page is always whole voteringar.
	if offset >= len(out) {
		return []ports.CommitteeVotering{}, total, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}
	return out[offset:end], total, nil
}

// forslagspunktKey identifies one decided förslagspunkt. Riksmöte is part of
// the key because beteckning numbering restarts each riksmöte: without it,
// NU1 punkt 2 in 2023/24 and NU1 punkt 2 in 2024/25 would be read as one
// twice-decided point when they are two separate decisions.
func forslagspunktKey(v ports.CommitteeVotering) string {
	return v.Riksmote + "\x00" + v.Beteckning + "\x00" + v.Forslagspunkt
}
