package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/curation/domain"
	"riksdagskollen/internal/curation/ports"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) InTx(ctx context.Context, fn func(ports.Store) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if err := fn(&store{tx: tx}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

const entryCols = `id, created_at, actor, on_behalf_of, chat_ref, action, entity, entity_id,
	old_value, new_value, undoes, request_key, request_hash`

func scanEntry(row pgx.Row) (*domain.Entry, error) {
	var e domain.Entry
	if err := row.Scan(&e.ID, &e.CreatedAt, &e.Actor, &e.OnBehalfOf, &e.ChatRef, &e.Action, &e.Entity,
		&e.EntityID, &e.OldValue, &e.NewValue, &e.Undoes, &e.RequestKey, &e.RequestHash); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) ListEntries(ctx context.Context, limit int, beforeID int64) ([]*domain.Entry, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+entryCols+` FROM aendringslogg WHERE ($2::bigint = 0 OR id < $2)
		   AND entity NOT IN ('koppling', 'motionkoppling', 'veckobatch')
		   AND NOT (entity = 'loefte'
		            AND coalesce(old_value->>'visibility','') NOT IN ('approved','archived')
		            AND coalesce(new_value->>'visibility','') NOT IN ('approved','archived'))
		 ORDER BY id DESC LIMIT $1`,
		limit, beforeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Entry{}
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

type store struct{ tx pgx.Tx }

func (s *store) LockRequest(ctx context.Context, key string) error {
	_, err := s.tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "aendringslogg:"+key)
	return err
}

func (s *store) EntryByKey(ctx context.Context, key string) (*domain.Entry, error) {
	e, err := scanEntry(s.tx.QueryRow(ctx, `SELECT `+entryCols+` FROM aendringslogg WHERE request_key = $1`, key))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return e, err
}

func (s *store) EntryByID(ctx context.Context, id int64) (*domain.Entry, error) {
	e, err := scanEntry(s.tx.QueryRow(ctx, `SELECT `+entryCols+` FROM aendringslogg WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return e, err
}

func (s *store) AppendEntry(ctx context.Context, e *domain.Entry) (*domain.Entry, error) {
	return scanEntry(s.tx.QueryRow(ctx,
		`INSERT INTO aendringslogg (actor, on_behalf_of, chat_ref, action, entity, entity_id,
		   old_value, new_value, undoes, request_key, request_hash)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING `+entryCols,
		e.Actor, e.OnBehalfOf, e.ChatRef, e.Action, e.Entity, e.EntityID,
		jsonOrNull(e.OldValue), jsonOrNull(e.NewValue), e.Undoes, e.RequestKey, e.RequestHash))
}

func jsonOrNull(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

const vmCols = `id, party, election_year, title, coalesce(to_char(published_on,'YYYY-MM-DD'),''),
	url, coalesce(pdf_url,''), state`

func scanVM(row pgx.Row) (*domain.Valmanifest, error) {
	var v domain.Valmanifest
	err := row.Scan(&v.ID, &v.Party, &v.ElectionYear, &v.Title, &v.PublishedOn, &v.URL, &v.PDFURL, &v.State)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *store) ValmanifestByKey(ctx context.Context, party string, year int, title string) (*domain.Valmanifest, error) {
	return scanVM(s.tx.QueryRow(ctx, `SELECT `+vmCols+` FROM valmanifest
		WHERE party=$1 AND election_year=$2 AND title=$3 FOR UPDATE`, party, year, title))
}

func (s *store) ValmanifestByID(ctx context.Context, id int64) (*domain.Valmanifest, error) {
	return scanVM(s.tx.QueryRow(ctx, `SELECT `+vmCols+` FROM valmanifest WHERE id=$1 FOR UPDATE`, id))
}

func pgCode(err error) string {
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return pe.Code
	}
	return ""
}

func (s *store) InsertValmanifest(ctx context.Context, v domain.Valmanifest) (*domain.Valmanifest, error) {
	var id any
	if v.ID != 0 {
		id = v.ID
	}
	// A savepoint keeps the outer transaction usable only until the caller
	// rolls back; every error here aborts the request anyway.
	out, err := scanVM(s.tx.QueryRow(ctx, `INSERT INTO valmanifest
		(id, party, election_year, title, published_on, url, pdf_url, state)
		VALUES (COALESCE($1::int, nextval('valmanifest_id_seq')), $2,$3,$4,NULLIF($5,'')::date,$6,NULLIF($7,''),$8)
		RETURNING `+vmCols, id, v.Party, v.ElectionYear, v.Title, v.PublishedOn, v.URL, v.PDFURL, v.State))
	switch pgCode(err) {
	case "23503":
		return nil, &domain.ValidationError{Msg: fmt.Sprintf("unknown party %q", v.Party)}
	case "23505":
		return nil, fmt.Errorf("%w: a Valmanifest with this key already exists", domain.ErrConflict)
	}
	return out, err
}

func (s *store) UpdateValmanifest(ctx context.Context, v domain.Valmanifest) error {
	_, err := s.tx.Exec(ctx, `UPDATE valmanifest SET party=$2, election_year=$3, title=$4,
		published_on=NULLIF($5,'')::date, url=$6, pdf_url=NULLIF($7,''), state=$8 WHERE id=$1`,
		v.ID, v.Party, v.ElectionYear, v.Title, v.PublishedOn, v.URL, v.PDFURL, v.State)
	switch pgCode(err) {
	case "23503":
		return &domain.ValidationError{Msg: fmt.Sprintf("unknown party %q", v.Party)}
	case "23505":
		return fmt.Errorf("%w: a Valmanifest with this key already exists", domain.ErrConflict)
	}
	return err
}

func (s *store) DeleteValmanifest(ctx context.Context, id int64) error {
	_, err := s.tx.Exec(ctx, `DELETE FROM valmanifest WHERE id=$1`, id)
	if pgCode(err) == "23503" {
		return fmt.Errorf("%w: other data still refers to this Valmanifest", domain.ErrConflict)
	}
	return err
}

const loefteCols = `id, party, valmanifest_id, text, coalesce(quote,''), coalesce(locator,''),
	coalesce(source_url,''), coalesce(topic,''), coalesce(specificity,''), utskott, mandate_period, visibility`

func scanLoefte(row pgx.Row) (*domain.Loefte, error) {
	var l domain.Loefte
	var vm *int64
	err := row.Scan(&l.ID, &l.Party, &vm, &l.Text, &l.Quote, &l.Locator, &l.SourceURL, &l.Topic,
		&l.Specificity, &l.Utskott, &l.MandatePeriod, &l.Visibility)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if vm != nil {
		l.ValmanifestID = *vm
	}
	if l.Utskott == nil {
		l.Utskott = []string{}
	}
	return &l, nil
}

func (s *store) LoefteByID(ctx context.Context, id int64) (*domain.Loefte, error) {
	return scanLoefte(s.tx.QueryRow(ctx, `SELECT `+loefteCols+` FROM loften WHERE id=$1 FOR UPDATE`, id))
}

func (s *store) LoefteByQuote(ctx context.Context, party string, vm int64, quote string) (*domain.Loefte, error) {
	return scanLoefte(s.tx.QueryRow(ctx, `SELECT `+loefteCols+` FROM loften
		WHERE party=$1 AND valmanifest_id=$2 AND quote=$3 AND visibility <> 'struck' ORDER BY id LIMIT 1`,
		party, vm, quote))
}

func (s *store) InsertLoefte(ctx context.Context, l domain.Loefte) (*domain.Loefte, error) {
	var id any
	if l.ID != 0 {
		id = l.ID
	}
	if l.Utskott == nil {
		l.Utskott = []string{}
	}
	// The Valmanifest must belong to the same party.
	var vmParty string
	err := s.tx.QueryRow(ctx, `SELECT party FROM valmanifest WHERE id=$1`, l.ValmanifestID).Scan(&vmParty)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &domain.ValidationError{Msg: fmt.Sprintf("unknown valmanifestId %d", l.ValmanifestID)}
	}
	if err != nil {
		return nil, err
	}
	if vmParty != l.Party {
		return nil, &domain.ValidationError{Msg: fmt.Sprintf("Valmanifest %d belongs to %s, not %s", l.ValmanifestID, vmParty, l.Party)}
	}
	out, err := scanLoefte(s.tx.QueryRow(ctx, `INSERT INTO loften
		(id, party, valmanifest_id, text, quote, locator, source_url, topic, specificity, utskott, mandate_period, visibility, approved_at)
		VALUES (COALESCE($1::int, nextval('loften_id_seq')), $2,$3,$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),$10,$11,$12,
		        CASE WHEN $12 = 'approved' THEN now() END)
		RETURNING `+loefteCols,
		id, l.Party, l.ValmanifestID, l.Text, l.Quote, l.Locator, l.SourceURL, l.Topic, l.Specificity,
		l.Utskott, l.MandatePeriod, l.Visibility))
	switch pgCode(err) {
	case "23503":
		return nil, &domain.ValidationError{Msg: fmt.Sprintf("unknown party %q", l.Party)}
	case "23514":
		return nil, &domain.ValidationError{Msg: "a Löfte needs a quote and a source URL"}
	}
	return out, err
}

func (s *store) UpdateLoefte(ctx context.Context, l domain.Loefte) error {
	if l.Utskott == nil {
		l.Utskott = []string{}
	}
	_, err := s.tx.Exec(ctx, `UPDATE loften SET text=$2, quote=NULLIF($3,''), locator=NULLIF($4,''),
		source_url=NULLIF($5,''), topic=NULLIF($6,''), specificity=NULLIF($7,''), utskott=$8, visibility=$9,
		approved_at = CASE WHEN $9 = 'approved' THEN coalesce(approved_at, now())
		                   WHEN $9 = 'proposed' THEN NULL ELSE approved_at END
		WHERE id=$1`,
		l.ID, l.Text, l.Quote, l.Locator, l.SourceURL, l.Topic, l.Specificity, l.Utskott, l.Visibility)
	if pgCode(err) == "23514" {
		return &domain.ValidationError{Msg: "a Löfte needs a quote and a source URL"}
	}
	return err
}

func (s *store) DeleteLoefte(ctx context.Context, id int64) error {
	_, err := s.tx.Exec(ctx, `DELETE FROM loften WHERE id=$1`, id)
	if pgCode(err) == "23503" {
		return fmt.Errorf("%w: other data still refers to this Löfte", domain.ErrConflict)
	}
	return err
}
