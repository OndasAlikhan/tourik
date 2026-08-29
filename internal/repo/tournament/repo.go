package tournament

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/OndasAlikhan/tourik/internal/config"
	tourDomain "github.com/OndasAlikhan/tourik/internal/domain/tournament"
	"github.com/OndasAlikhan/tourik/internal/repo/outbox"
)

type Repo struct {
	db         *sqlx.DB
	config     config.AppConfig
	outboxRepo outbox.Repo
}

func New(db *sqlx.DB, outboxRepo outbox.Repo, config config.AppConfig) Repo {
	return Repo{
		db:         db,
		outboxRepo: outboxRepo,
		config:     config,
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
		insert into tournaments (name, discipline, status, bracket_format, max_participants, start_date, end_date, created_at, updated_at)
		values (:name, :discipline, :status, :bracket_format, :max_participants, :start_date, :end_date, now(), now())
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
		set name = :name, discipline = :discipline, status = :status, bracket_format = :bracket_format,
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

// CreateWithEvent inserts the tournament and an outbox event in a single
// transaction. buildEvent receives the row as it exists in the DB (with its
// generated ID) so the event key/payload can reference it.
func (r Repo) CreateWithEvent(ctx context.Context, t Tournament) (Tournament, error) {
	const op = "tournament.Repo.CreateWithEvent"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	const query = `
		insert into tournaments (name, discipline, status, bracket_format, max_participants, start_date, end_date, created_at, updated_at)
		values (:name, :discipline, :status, :bracket_format, :max_participants, :start_date, :end_date, now(), now())
		returning *`

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: prepare named query: %w", op, err)
	}

	var result Tournament
	err = stmt.GetContext(ctx, &result, t)
	stmt.Close()
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: error inserting tournament: %w", op, err)
	}

	event, err := r.buildEvent(result)
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: could not build outbox.Event", op)
	}
	if err := r.outboxRepo.Insert(ctx, tx, event); err != nil {
		return Tournament{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return Tournament{}, fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return result, nil
}

func (r Repo) UpdateWithEvent(ctx context.Context, t Tournament) (Tournament, error) {
	const op = "tournament.Repo.UpdateWithEvent"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	const query = `
		update tournaments
		set name = :name, discipline = :discipline, status = :status, bracket_format = :bracket_format,
			max_participants = :max_participants, start_date = :start_date, end_date = :end_date, updated_at = now()
		where id = :id
		returning *`

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: prepare named query: %w", op, err)
	}

	var result Tournament
	err = stmt.GetContext(ctx, &result, t)
	stmt.Close()
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: error updating tournament: %w", op, err)
	}

	event, err := r.buildEvent(result)
	if err != nil {
		return Tournament{}, fmt.Errorf("%s: could not build outbox.Event", op)
	}

	if err := r.outboxRepo.Insert(ctx, tx, event); err != nil {
		return Tournament{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return Tournament{}, fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return result, nil
}

func (r Repo) buildEvent(tour Tournament) (outbox.Event, error) {
	const op = "repo.tournament.buildEvent"
	payload, err := json.Marshal(tour)
	if err != nil {
		return outbox.Event{}, fmt.Errorf("%s: could not marshal payload", op)
	}
	msg, err := json.Marshal(tourDomain.TournamentEventMessage{
		EventID:   uuid.New(),
		EventType: tourDomain.TournamentCreated,
		OccuredAt: time.Now(),
		Payload:   payload,
	})
	if err != nil {
		return outbox.Event{}, fmt.Errorf("%s: could not marshal msg", op)
	}
	return outbox.Event{
		Topic:     r.config.OutboxTopic,
		Key:       strconv.Itoa(tour.ID),
		Message:   msg,
		CreatedAt: time.Now(),
		SentAt:    nil,
	}, nil
}
