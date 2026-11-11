package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"riksdagskollen/internal/curation/domain"
)

func (s *store) motionDokID(ctx context.Context, r domain.MotionRef) (string, error) {
	var id string
	err := s.tx.QueryRow(ctx, `SELECT dok_id FROM motions WHERE riksmote=$1 AND beteckning=$2`, r.Riksmote, r.Beteckning).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", &domain.ValidationError{Msg: fmt.Sprintf("unknown motion %s:%s", r.Riksmote, r.Beteckning)}
	}
	return id, err
}

func (s *store) hasYrkande(ctx context.Context, dokID string, nr int) (bool, error) {
	var ok bool
	err := s.tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM motion_yrkanden WHERE dok_id=$1 AND nummer=$2)`, dokID, nr).Scan(&ok)
	return ok, err
}

func (s *store) ResolveMotionSuggestion(ctx context.Context, sg domain.MotionSuggestion) (*domain.MotionKoppling, error) {
	k := &domain.MotionKoppling{Kind: sg.Kind}
	var err error
	if k.MotionDokID, err = s.motionDokID(ctx, sg.Motion); err != nil {
		return nil, err
	}
	switch sg.Kind {
	case domain.KindFraga:
		k.Rule = domain.RuleProseCitation
		q := sg.Question
		err := s.tx.QueryRow(ctx, `SELECT dok_id FROM questions WHERE kind=$1 AND riksmote=$2 AND beteckning=$3`,
			q.Kind, q.Riksmote, q.Beteckning).Scan(&k.QuestionDokID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.ValidationError{Msg: fmt.Sprintf("unknown %s %s:%s", q.Kind, q.Riksmote, q.Beteckning)}
		}
		if err != nil {
			return nil, err
		}
	case domain.KindKrav:
		k.Rule = domain.RuleSimilarWording
		k.Yrkande = sg.Motion.Yrkande
		if k.OtherMotionID, err = s.motionDokID(ctx, *sg.Other); err != nil {
			return nil, err
		}
		k.OtherYrkande = sg.Other.Yrkande
		for _, c := range []struct {
			id string
			nr int
		}{{k.MotionDokID, k.Yrkande}, {k.OtherMotionID, k.OtherYrkande}} {
			if ok, err := s.hasYrkande(ctx, c.id, c.nr); err != nil {
				return nil, err
			} else if !ok {
				return nil, &domain.ValidationError{Msg: fmt.Sprintf("motion %s has no yrkande %d", c.id, c.nr)}
			}
		}
	}
	return k, nil
}

const mkCols = `id, kind, rule, state, coalesce(question_dok_id,''), motion_dok_id, coalesce(yrkande_nr,0),
	coalesce(other_motion_dok_id,''), coalesce(other_yrkande_nr,0),
	confidence::float8, threshold::float8, reasoning, origin`

func scanMK(row pgx.Row) (*domain.MotionKoppling, error) {
	var k domain.MotionKoppling
	if err := row.Scan(&k.ID, &k.Kind, &k.Rule, &k.State, &k.QuestionDokID, &k.MotionDokID, &k.Yrkande,
		&k.OtherMotionID, &k.OtherYrkande, &k.Confidence, &k.Threshold, &k.Reasoning, &k.Origin); err != nil {
		return nil, err
	}
	return &k, nil
}

func (s *store) MotionKopplingByPair(ctx context.Context, k domain.MotionKoppling) (*domain.MotionKoppling, error) {
	var row pgx.Row
	if k.Kind == domain.KindFraga {
		row = s.tx.QueryRow(ctx, `SELECT `+mkCols+` FROM motion_kopplingar
			WHERE kind='fraga' AND question_dok_id=$1 AND motion_dok_id=$2 FOR UPDATE`, k.QuestionDokID, k.MotionDokID)
	} else {
		// The pair is unordered: look for either order.
		row = s.tx.QueryRow(ctx, `SELECT `+mkCols+` FROM motion_kopplingar
			WHERE kind='krav' AND ((motion_dok_id=$1 AND yrkande_nr=$2 AND other_motion_dok_id=$3 AND other_yrkande_nr=$4)
			   OR (motion_dok_id=$3 AND yrkande_nr=$4 AND other_motion_dok_id=$1 AND other_yrkande_nr=$2)) FOR UPDATE`,
			k.MotionDokID, k.Yrkande, k.OtherMotionID, k.OtherYrkande)
	}
	out, err := scanMK(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return out, err
}

func (s *store) InsertMotionKoppling(ctx context.Context, k domain.MotionKoppling) (*domain.MotionKoppling, error) {
	err := s.tx.QueryRow(ctx, `
		INSERT INTO motion_kopplingar (kind, rule, state, question_dok_id, motion_dok_id, yrkande_nr,
			other_motion_dok_id, other_yrkande_nr, confidence, threshold, reasoning, origin)
		VALUES ($1,$2,$3,NULLIF($4,''),$5,NULLIF($6,0),NULLIF($7,''),NULLIF($8,0),$9,$10,$11,$12) RETURNING id`,
		k.Kind, k.Rule, k.State, k.QuestionDokID, k.MotionDokID, k.Yrkande, k.OtherMotionID, k.OtherYrkande,
		k.Confidence, k.Threshold, k.Reasoning, k.Origin).Scan(&k.ID)
	return &k, err
}

func (s *store) MotionKopplingByID(ctx context.Context, id int64) (*domain.MotionKoppling, error) {
	k, err := scanMK(s.tx.QueryRow(ctx, `SELECT `+mkCols+` FROM motion_kopplingar WHERE id=$1 FOR UPDATE`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var bid int64
	var nr int
	var ans *string
	err = s.tx.QueryRow(ctx, `SELECT veckobatch_id, nr, answer FROM veckobatch_items
		WHERE motion_koppling_id=$1 ORDER BY veckobatch_id DESC LIMIT 1`, id).Scan(&bid, &nr, &ans)
	if err == nil {
		k.VeckobatchID, k.Nr = &bid, &nr
		if ans != nil {
			k.Answer = *ans
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return k, nil
}

func (s *store) DecideMotionItem(ctx context.Context, batchID int64, nr int, id int64, state, answer string) error {
	if _, err := s.tx.Exec(ctx, `UPDATE motion_kopplingar SET state=$2,
		decided_at = CASE WHEN $2 IN ('bekraftad','avvisad') THEN now() ELSE NULL END WHERE id=$1`, id, state); err != nil {
		return err
	}
	_, err := s.tx.Exec(ctx, `UPDATE veckobatch_items SET answer = NULLIF($4,''),
		answered_at = CASE WHEN $4 = '' THEN NULL ELSE now() END
		WHERE veckobatch_id=$1 AND nr=$2 AND motion_koppling_id=$3`, batchID, nr, id, answer)
	return err
}

// loadMotionItems reads the motion-link items of a batch, with what the
// Redaktör needs to judge them in chat.
func loadMotionItems(ctx context.Context, q querier, batchID int64, lockK string) ([]domain.BatchItem, error) {
	rows, err := q.Query(ctx, `
		SELECT i.nr, k.id, k.kind, k.rule, k.state,
		       EXISTS (SELECT 1 FROM veckobatch_items o WHERE o.motion_koppling_id = k.id AND o.veckobatch_id < i.veckobatch_id),
		       k.confidence::float8, k.reasoning, i.answer,
		       coalesce(qq.kind,''), coalesce(qq.riksmote,''), coalesce(qq.beteckning,''), coalesce(qq.title,''),
		       coalesce(qq.asker_name || ' (' || qq.asker_party || ')',''), left(coalesce(qq.body_text,''), 600), coalesce(qq.dok_id,''),
		       m.dok_id, m.riksmote, m.beteckning, m.title, `+partiesSQL("m")+`,
		       coalesce(k.yrkande_nr,0), coalesce((SELECT y.wording FROM motion_yrkanden y WHERE y.dok_id=m.dok_id AND y.nummer=k.yrkande_nr),''),
		       coalesce(o.dok_id,''), coalesce(o.riksmote,''), coalesce(o.beteckning,''), coalesce(o.title,''),
		       `+partiesSQL("o")+`,
		       coalesce(k.other_yrkande_nr,0), coalesce((SELECT y.wording FROM motion_yrkanden y WHERE y.dok_id=o.dok_id AND y.nummer=k.other_yrkande_nr),'')
		  FROM veckobatch_items i
		  JOIN motion_kopplingar k ON k.id = i.motion_koppling_id
		  JOIN motions m ON m.dok_id = k.motion_dok_id
		  LEFT JOIN motions o ON o.dok_id = k.other_motion_dok_id
		  LEFT JOIN questions qq ON qq.dok_id = k.question_dok_id
		 WHERE i.veckobatch_id = $1
		 ORDER BY i.nr`+lockK, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.BatchItem
	for rows.Next() {
		var it domain.BatchItem
		mi := domain.MotionBatchItem{}
		var q domain.QuestionView
		var qDok, oDok string
		var o domain.MotionView
		if err := rows.Scan(&it.Nr, &it.MotionKopplingID, &mi.Kind, &mi.Rule, &it.State, &it.RolledOver,
			&it.Confidence, &it.Reasoning, &it.Answer,
			&q.Kind, &q.Riksmote, &q.Beteckning, &q.Title, &q.Asker, &q.Excerpt, &qDok,
			&mi.Motion.SourceURL, &mi.Motion.Riksmote, &mi.Motion.Beteckning, &mi.Motion.Title, &mi.Motion.Parties,
			&mi.Motion.Yrkande, &mi.Motion.Wording,
			&oDok, &o.Riksmote, &o.Beteckning, &o.Title, &o.Parties, &o.Yrkande, &o.Wording); err != nil {
			return nil, err
		}
		mi.Motion.SourceURL = sourceURL(mi.Motion.SourceURL)
		if qDok != "" {
			q.SourceURL = sourceURL(qDok)
			mi.Question = &q
		}
		if oDok != "" {
			o.SourceURL = sourceURL(oDok)
			mi.Other = &o
		}
		it.Kind, it.Motion = mi.Kind, &mi
		out = append(out, it)
	}
	return out, rows.Err()
}

func sourceURL(dokID string) string { return "https://data.riksdagen.se/dokument/" + dokID }

// partiesSQL is the distinct signer parties of motion alias a, in signing order.
func partiesSQL(a string) string {
	return `COALESCE((SELECT array_agg(s.party ORDER BY s.first) FROM (
		SELECT x.party, min(x.ordning) AS first FROM motion_authors x
		 WHERE x.dok_id = ` + a + `.dok_id AND x.party NOT IN ('', '-') GROUP BY x.party) s), '{}')`
}
