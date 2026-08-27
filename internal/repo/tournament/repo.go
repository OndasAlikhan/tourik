package tournament

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repo {
	return Repo{
		db: db,
	}
}

func (r Repo) ByID(ctx context.Context, id int) (Tournament, error) {
	const op = "tournament.Repo.ByID"

	const query = `select * from tournaments where id = $1`

	var result Tournament
	if err := r.db.GetContext(ctx, &result, query, id); err != nil {
		return Tournament{}, fmt.Errorf("%s: error on query: %w", op, err)
	}

	return result, nil
}

func (r Repo) List(ctx context.Context) ([]Tournament, error) {
	const op = "tournament.Repo.List"

	const query = `select * from tournaments`

	result := []Tournament{}
	if err := r.db.SelectContext(ctx, &result, query); err != nil {
		return []Tournament{}, fmt.Errorf("%s: error on db query: %w", op, err)
	}
	return result, nil
}

func (r Repo) Create(ctx context.Context, t Tournament) (Tournament, error) {
	const op = "tournament.Repo.Create"

	const query = `
		insert into tournaments (name, discipline, bracket_format, max_participants, start_date, end_date, created_at, updated_at)
		values (:name, :discipline, :bracket_format, :max_participants, :start_date, :end_date, now(), now())
		returning *`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: prepare named query: %w", op, err)
	}
	defer stmt.Close()

	var result Tournament
	if err := stmt.GetContext(ctx, &result, t); err != nil {
		return Tournament{}, fmt.Errorf("%s: error inserting tournament: %w", op, err)
	}
	return result, nil
}

func (r Repo) Update(ctx context.Context, t Tournament) (Tournament, error) {
	const op = "tournament.Repo.Update"

	const query = `
		update tournaments
		set name = :name, discipline = :discipline, bracket_format = :bracket_format,
			max_participants = :max_participants, start_date = :start_date, end_date = :end_date, updated_at = now()
		where id = :id
		returning *`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: prepare named query: %w", op, err)
	}
	defer stmt.Close()

	var result Tournament
	if err := stmt.GetContext(ctx, &result, t); err != nil {
		return Tournament{}, fmt.Errorf("%s: error updating tournament: %w", op, err)
	}
	return result, nil
}
