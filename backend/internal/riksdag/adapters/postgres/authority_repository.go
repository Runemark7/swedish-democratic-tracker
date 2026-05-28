package postgres

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"riksdagskollen/internal/riksdag/domain"
	"riksdagskollen/internal/riksdag/ports"
)

type AuthorityRepository struct {
	pool *pgxpool.Pool
}

func NewAuthorityRepository(pool *pgxpool.Pool) *AuthorityRepository {
	return &AuthorityRepository{pool: pool}
}

func (r *AuthorityRepository) UpsertAuthorities(ctx context.Context, items []domain.RegisteredAuthority) (int, error) {
	n := 0
	for _, a := range items {
		// Register fields only (Phase 1). Expenditure/headcount columns are left
		// untouched on update so later enrichment is not wiped.
		_, err := r.pool.Exec(ctx, `
			INSERT INTO authorities
				(org_number, slug, name, type, principal_body, department,
				 under_government, website, sfs, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, now())
			ON CONFLICT (org_number) DO UPDATE SET
				slug             = EXCLUDED.slug,
				name             = EXCLUDED.name,
				type             = EXCLUDED.type,
				principal_body   = EXCLUDED.principal_body,
				under_government = EXCLUDED.under_government,
				website          = EXCLUDED.website,
				sfs              = EXCLUDED.sfs,
				updated_at       = now()
		`, a.OrgNumber, a.Slug, a.Name, a.Type, a.PrincipalBody, a.Department,
			a.UnderGovernment, a.Website, a.SFS)
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (r *AuthorityRepository) List(ctx context.Context, f ports.AuthorityFilter) ([]domain.RegisteredAuthority, error) {
	where, args := buildWhere(f)
	sql := `
		SELECT org_number, slug, name, type, principal_body, department,
		       under_government, website, sfs, expenditure_mdkr, budget_mdkr,
		       headcount_int, year, updated_at
		FROM authorities ` + where + `
		ORDER BY expenditure_mdkr DESC NULLS LAST, name ASC`
	if f.Limit > 0 {
		args = append(args, f.Limit)
		sql += " LIMIT $" + strconv.Itoa(len(args))
		args = append(args, f.Offset)
		sql += " OFFSET $" + strconv.Itoa(len(args))
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.RegisteredAuthority
	for rows.Next() {
		var a domain.RegisteredAuthority
		if err := rows.Scan(
			&a.OrgNumber, &a.Slug, &a.Name, &a.Type, &a.PrincipalBody, &a.Department,
			&a.UnderGovernment, &a.Website, &a.SFS, &a.ExpenditureMdkr, &a.BudgetMdkr,
			&a.HeadcountInt, &a.Year, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AuthorityRepository) GetBySlug(ctx context.Context, slug string) (*domain.RegisteredAuthority, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT org_number, slug, name, type, principal_body, department,
		       under_government, website, sfs, expenditure_mdkr, budget_mdkr,
		       headcount_int, year, updated_at
		FROM authorities WHERE slug = $1
	`, slug)
	var a domain.RegisteredAuthority
	if err := row.Scan(
		&a.OrgNumber, &a.Slug, &a.Name, &a.Type, &a.PrincipalBody, &a.Department,
		&a.UnderGovernment, &a.Website, &a.SFS, &a.ExpenditureMdkr, &a.BudgetMdkr,
		&a.HeadcountInt, &a.Year, &a.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AuthorityRepository) Count(ctx context.Context, f ports.AuthorityFilter) (int, error) {
	where, args := buildWhere(f)
	var n int
	err := r.pool.QueryRow(ctx, "SELECT count(*) FROM authorities "+where, args...).Scan(&n)
	return n, err
}

func (r *AuthorityRepository) UpdateExpenditure(ctx context.Context, items []ports.ExpenditureUpdate) (int, int, error) {
	matched, unmatched := 0, 0
	for _, e := range items {
		historyJSON, err := json.Marshal(e.History)
		if err != nil {
			return matched, unmatched, err
		}
		tag, err := r.pool.Exec(ctx, `
			UPDATE authorities
			SET expenditure_mdkr    = $2,
			    budget_mdkr         = $3,
			    year                = GREATEST(year, $4),
			    expenditure_history = $5,
			    updated_at          = now()
			WHERE org_number = $1
		`, e.OrgNumber, e.ExpenditureMdkr, e.BudgetMdkr, e.Year, historyJSON)
		if err != nil {
			return matched, unmatched, err
		}
		if tag.RowsAffected() == 1 {
			matched++
		} else {
			unmatched++
		}
	}
	return matched, unmatched, nil
}

func (r *AuthorityRepository) UpdateEnrichment(ctx context.Context, items []ports.Enrichment) (int, int, error) {
	matched, unmatched := 0, 0
	for _, e := range items {
		historyJSON, err := json.Marshal(e.History)
		if err != nil {
			return matched, unmatched, err
		}
		tag, err := r.pool.Exec(ctx, `
			UPDATE authorities
			SET department        = CASE WHEN $2 <> '' THEN $2 ELSE department END,
			    headcount_int     = $3,
			    year              = GREATEST(year, $4),
			    headcount_history = $5,
			    updated_at        = now()
			WHERE org_number = $1
		`, e.OrgNumber, e.Department, e.HeadcountInt, e.Year, historyJSON)
		if err != nil {
			return matched, unmatched, err
		}
		if tag.RowsAffected() == 1 {
			matched++
		} else {
			unmatched++
		}
	}
	return matched, unmatched, nil
}

func buildWhere(f ports.AuthorityFilter) (string, []any) {
	var conds []string
	var args []any
	if q := strings.TrimSpace(f.Query); q != "" {
		args = append(args, "%"+q+"%")
		conds = append(conds, "name ILIKE $"+strconv.Itoa(len(args)))
	}
	if f.UnderGovernment != nil {
		args = append(args, *f.UnderGovernment)
		conds = append(conds, "under_government = $"+strconv.Itoa(len(args)))
	}
	if len(conds) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(conds, " AND "), args
}
