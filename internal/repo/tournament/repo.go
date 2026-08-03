package tournament

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repo {
	return Repo{
		db: db,
	}
}

func (r Repo) List(ctx context.Context) ([]Tournament, error) {
	const op = "tournament.Repo.List"

	const query = `select * from tournaments`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []Tournament{}, nil
		}
		return []Tournament{}, fmt.Errorf("%s: error on db query : %w", op, err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[Tournament])
	if err != nil {
		return []Tournament{}, fmt.Errorf("%s: collect rows: %w", op, err)
	}
	return result, nil
}
