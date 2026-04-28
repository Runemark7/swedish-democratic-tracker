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
	"AU":   "Arbetsmarknad",
	"CU":   "Bostäder",
	"FiU":  "Ekonomi",
	"FöU":  "Försvar",
	"JuU":  "Rättsväsen",
	"KrU":  "Kultur",
	"KU":   "Demokrati",
	"MJU":  "Miljö",
	"NU":   "Näringsliv",
	"SfU":  "Välfärd",
	"SkU":  "Skatt",
	"SoU":  "Hälsa",
	"TU":   "Infrastruktur",
	"UbU":  "Utbildning",
	"UU":   "Utrikes",
	"KE":   "Konstitution",
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
	const q = `
		SELECT
			beteckning,
			forslagspunkt,
			MAX(COALESCE(NULLIF(document_title, ''), beteckning)) AS title,
			MAX(created_at) AS latest_at,
			COUNT(CASE WHEN vote_result = 'Ja'  THEN 1 END) AS ja_count,
			COUNT(CASE WHEN vote_result = 'Nej' THEN 1 END) AS nej_count
		FROM votes
		GROUP BY beteckning, forslagspunkt
		ORDER BY latest_at DESC
		LIMIT $1
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
