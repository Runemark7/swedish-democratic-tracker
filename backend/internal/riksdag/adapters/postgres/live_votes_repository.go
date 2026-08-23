package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/riksdag/domain"
)

type LiveVotesRepository struct {
	db *pgxpool.Pool
}

func NewLiveVotesRepository(db *pgxpool.Pool) *LiveVotesRepository {
	return &LiveVotesRepository{db: db}
}

// betecknigToTag maps a riksdag committee code prefix to a display tag.
var beteckningToTag = map[string]string{
	"AU":  "Arbetsmarknad",
	"CU":  "Bostäder",
	"FiU": "Ekonomi",
	"FöU": "Försvar",
	"JuU": "Rättsväsen",
	"KrU": "Kultur",
	"KU":  "Demokrati",
	"MJU": "Miljö",
	"NU":  "Näringsliv",
	"SfU": "Välfärd",
	"SkU": "Skatt",
	"SoU": "Hälsa",
	"TU":  "Infrastruktur",
	"UbU": "Utbildning",
	"UU":  "Utrikes",
	"KE":  "Konstitution",
}

func tagFromBeteckning(b string) string {
	// Beteckning format: "FiU2425:20" — prefix is the letters before the digits
	prefix := strings.TrimRightFunc(b, func(r rune) bool {
		return r >= '0' && r <= '9' || r == ':' || r == '/'
	})
	// Strip trailing digits from the trimmed result (e.g. "FöU2425" → "FöU")
	i := 0
	for i < len(prefix) && (prefix[i] < '0' || prefix[i] > '9') {
		i++
	}
	key := prefix[:i]
	if tag, ok := beteckningToTag[key]; ok {
		return tag
	}
	return "Riksdagen"
}

func (r *LiveVotesRepository) ListLiveVotes(ctx context.Context, limit int) ([]domain.LiveVote, error) {
	// The ranking is settled before any ballot is read. title and latest_at are
	// votering-level, so the newest points are chosen from a 15 835-row table
	// and only those points' ballots are counted -- rather than aggregating all
	// 5.5 million to return $1 rows.
	const q = `
		WITH recent AS (
			SELECT
				beteckning,
				forslagspunkt,
				MAX(COALESCE(NULLIF(document_title, ''), beteckning)) AS title,
				MAX(created_at) AS latest_at
			FROM voteringar
			GROUP BY beteckning, forslagspunkt
			ORDER BY latest_at DESC
			LIMIT $1
		)
		SELECT
			r.beteckning,
			r.forslagspunkt,
			r.title,
			r.latest_at,
			COUNT(*) FILTER (WHERE b.vote_result = 'Ja')  AS ja_count,
			COUNT(*) FILTER (WHERE b.vote_result = 'Nej') AS nej_count
		FROM recent r
		JOIN voteringar vg
		  ON vg.beteckning = r.beteckning AND vg.forslagspunkt = r.forslagspunkt
		JOIN ballots b ON b.votering_ref = vg.id
		GROUP BY r.beteckning, r.forslagspunkt, r.title, r.latest_at
		ORDER BY r.latest_at DESC
	`
	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []domain.LiveVote
	for rows.Next() {
		var v domain.LiveVote
		var forslagspunkt string
		if err := rows.Scan(&v.Beteckning, &forslagspunkt, &v.Title, &v.Date, &v.JaCount, &v.NejCount); err != nil {
			return nil, err
		}
		if v.JaCount > v.NejCount {
			v.Status = "Bifall"
		} else {
			v.Status = "Avslag"
		}
		v.Margin = fmt.Sprintf("%d–%d", v.JaCount, v.NejCount)
		v.Tag = tagFromBeteckning(v.Beteckning)
		votes = append(votes, v)
	}
	return votes, rows.Err()
}
