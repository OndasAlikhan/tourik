package participant

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	sq "github.com/Masterminds/squirrel"
	"github.com/OndasAlikhan/tourik/internal/config"
	"github.com/OndasAlikhan/tourik/internal/domain"
	tourDomain "github.com/OndasAlikhan/tourik/internal/domain/tournament"
	"github.com/OndasAlikhan/tourik/internal/repo/outbox"
	"github.com/jmoiron/sqlx"
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

func (r Repo) ByID(ctx context.Context, id int) (Participant, error) {
	const op = "participant.Repo.ByID"

	const query = `select * from participants where id = $1`

	var result Participant
	if err := r.db.GetContext(ctx, &result, query, id); err != nil {
		return Participant{}, fmt.Errorf("%s: error on query: %w", op, err)
	}

	return result, nil
}

func (r Repo) List(ctx context.Context, filter domain.ParticipantFilter) ([]Participant, error) {
	const op = "participant.Repo.List"

	builder := sq.Select("participants.*").
		From("participants").
		PlaceholderFormat(sq.Dollar)

	if filter.TournamentID != 0 {
		builder = builder.
			Join("tournament_participants tp ON tp.participant_id = participants.id").
			Where(sq.Eq{"tp.tournament_id": filter.TournamentID})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return []Participant{}, fmt.Errorf("%s: build query: %w", op, err)
	}

	result := []Participant{}
	if err := r.db.SelectContext(ctx, &result, query, args...); err != nil {
		return []Participant{}, fmt.Errorf("%s: error on db query: %w", op, err)
	}
	return result, nil
}

func (r Repo) Create(ctx context.Context, p Participant) (Participant, error) {
	const op = "participant.Repo.Create"

	const query = `
		insert into participants (name, created_at, updated_at)
		values (:name, now(), now())
		returning *`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return Participant{}, fmt.Errorf("%s: prepare named query: %w", op, err)
	}
	defer stmt.Close()

	var result Participant
	if err := stmt.GetContext(ctx, &result, p); err != nil {
		return Participant{}, fmt.Errorf("%s: error inserting participant: %w", op, err)
	}
	return result, nil
}

func (r Repo) Update(ctx context.Context, p Participant) (Participant, error) {
	const op = "participant.Repo.Update"

	const query = `
		update participants
		set name = :name, updated_at = now()
		where id = :id
		returning *`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return Participant{}, fmt.Errorf("%s: prepare named query: %w", op, err)
	}
	defer stmt.Close()

	var result Participant
	if err := stmt.GetContext(ctx, &result, p); err != nil {
		return Participant{}, fmt.Errorf("%s: error updating participant: %w", op, err)
	}
	return result, nil
}

func (r Repo) CreateForTournament(ctx context.Context, tournamentID int, p Participant) (Participant, error) {
	const op = "participant.Repo.CreateForTournament"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return Participant{}, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	const insertParticipant = `
		insert into participants (name, created_at, updated_at)
		values (:name, now(), now())
		returning *`

	stmt, err := tx.PrepareNamedContext(ctx, insertParticipant)
	if err != nil {
		return Participant{}, fmt.Errorf("%s: prepare named query: %w", op, err)
	}

	var result Participant
	err = stmt.GetContext(ctx, &result, p)
	stmt.Close()
	if err != nil {
		return Participant{}, fmt.Errorf("%s: error inserting participant: %w", op, err)
	}

	const insertLink = `insert into tournament_participants (tournament_id, participant_id) values ($1, $2)`
	if _, err := tx.ExecContext(ctx, insertLink, tournamentID, result.ID); err != nil {
		return Participant{}, fmt.Errorf("%s: error linking participant to tournament: %w", op, err)
	}

	event, err := r.buildEvent(result)
	if err != nil {
		return Participant{}, fmt.Errorf("%s: build outbox event: %w", op, err)
	}

	if err := r.outboxRepo.Insert(ctx, tx, event); err != nil {
		return Participant{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return Participant{}, fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return result, nil
}

func (r Repo) DeleteFromTournament(ctx context.Context, tournamentID, participantID int) error {
	const op = "participant.Repo.DeleteFromTournament"

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	const deleteLink = `delete from tournament_participants where tournament_id = $1 and participant_id = $2`
	res, err := tx.ExecContext(ctx, deleteLink, tournamentID, participantID)
	if err != nil {
		return fmt.Errorf("%s: error unlinking participant: %w", op, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: error checking rows affected: %w", op, err)
	}
	if affected == 0 {
		return fmt.Errorf("%s: %w", op, sql.ErrNoRows)
	}

	const deleteParticipant = `delete from participants where id = $1`
	if _, err := tx.ExecContext(ctx, deleteParticipant, participantID); err != nil {
		return fmt.Errorf("%s: error deleting participant: %w", op, err)
	}

	event, err := r.buildDeleteEvent(participantID, tournamentID)
	if err != nil {
		return fmt.Errorf("%s: build outbox event: %w", op, err)
	}

	if err := r.outboxRepo.Insert(ctx, tx, event); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return nil
}

func (r Repo) buildEvent(pt Participant) (outbox.Event, error) {
	const op = "repo.tournament.buildEvent"
	payload, err := json.Marshal(pt)
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
		Key:       strconv.Itoa(pt.ID),
		Message:   msg,
		CreatedAt: time.Now(),
		SentAt:    nil,
	}, nil
}

func (r Repo) buildDeleteEvent(ptID int, tourID int) (outbox.Event, error) {
	const op = "repo.tournament.buildEvent"
	payload, err := json.Marshal(map[string]int{
		"participantID": ptID,
		"tournamentID":  tourID,
	})
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
		Key:       strconv.Itoa(ptID),
		Message:   msg,
		CreatedAt: time.Now(),
		SentAt:    nil,
	}, nil
}
