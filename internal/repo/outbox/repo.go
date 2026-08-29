package outbox

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

func (r Repo) Insert(ctx context.Context, tx *sqlx.Tx, event Event) error {
	const op = "outbox.Insert"

	const query = `insert into outbox (topic, key, message) values (:topic, :key, :message)`

	if _, err := tx.NamedExecContext(ctx, query, map[string]any{
		"topic":   event.Topic,
		"key":     event.Key,
		"message": event.Message,
	}); err != nil {
		return fmt.Errorf("%s: error inserting outbox event: %w", op, err)
	}
	return nil
}

func (r Repo) ListPending(ctx context.Context, limit int) ([]Event, error) {
	const op = "outbox.Repo.ListPending"

	const query = `select * from outbox where sent_at is null order by created_at limit $1`

	result := []Event{}
	if err := r.db.SelectContext(ctx, &result, query, limit); err != nil {
		return []Event{}, fmt.Errorf("%s: error on db query: %w", op, err)
	}
	return result, nil
}

func (r Repo) MarkSent(ctx context.Context, id int) error {
	const op = "outbox.Repo.MarkSent"

	const query = `update outbox set sent_at = now() where id = $1`

	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("%s: error updating outbox event: %w", op, err)
	}
	return nil
}
