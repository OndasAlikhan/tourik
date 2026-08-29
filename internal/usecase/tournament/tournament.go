package tournament

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/OndasAlikhan/tourik/internal/domain"
	tourDomain "github.com/OndasAlikhan/tourik/internal/domain/tournament"
	"github.com/OndasAlikhan/tourik/internal/repo/participant"
	"github.com/OndasAlikhan/tourik/internal/repo/tournament"
)

type Repo interface {
	ByID(ctx context.Context, id int) (tournament.Tournament, error)
	List(ctx context.Context) ([]tournament.Tournament, error)
	Create(ctx context.Context, t tournament.Tournament) (tournament.Tournament, error)
	Update(ctx context.Context, t tournament.Tournament) (tournament.Tournament, error)
	CreateWithEvent(ctx context.Context, t tournament.Tournament) (tournament.Tournament, error)
	UpdateWithEvent(ctx context.Context, t tournament.Tournament) (tournament.Tournament, error)
}

type tournamentEventPayload struct {
	EventType    string    `json:"event_type"`
	OccurredAt   time.Time `json:"occurred_at"`
	TournamentID int       `json:"tournament_id"`
	Status       string    `json:"status"`
}
type ParticipantRepo interface {
	List(ctx context.Context, filter domain.ParticipantFilter) ([]participant.Participant, error)
}
type Usecase struct {
	repo            Repo
	participantRepo ParticipantRepo
}

func New(repo tournament.Repo, participantRepo participant.Repo) Usecase {
	return Usecase{
		repo:            repo,
		participantRepo: participantRepo,
	}
}

func (uc Usecase) List(ctx context.Context) (result []tourDomain.Tournament, err error) {
	const op = "usecase.tournament.List"
	tournaments, err := uc.repo.List(ctx)
	if err != nil {
		return []tourDomain.Tournament{}, fmt.Errorf("%s: error getting tournament list: %w", op, err)
	}

	for _, val := range tournaments {
		result = append(result, val.ToDomain())
	}

	return result, nil
}

func (uc Usecase) ByID(ctx context.Context, id int) (tourDomain.Tournament, error) {
	const op = "usecase.tournament.ByID"

	result, err := uc.repo.ByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tourDomain.Tournament{}, fmt.Errorf("%s: %w", op, tourDomain.ErrTourNotFound)
		}
		return tourDomain.Tournament{}, fmt.Errorf("%s: error getting tournament: %w", op, err)
	}

	return result.ToDomain(), nil
}

func (uc Usecase) Create(ctx context.Context, t tourDomain.Tournament) (tourDomain.Tournament, error) {
	const op = "usecase.tournament.Create"

	if t.Status == "" {
		t.Status = tourDomain.Draft
	}

	created, err := uc.repo.CreateWithEvent(ctx, tournament.FromDomain(t))
	if err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: error creating tournament: %w", op, err)
	}

	return created.ToDomain(), nil
}

func (uc Usecase) Update(ctx context.Context, t tourDomain.Tournament) (tourDomain.Tournament, error) {
	const op = "usecase.tournament.Update"

	updated, err := uc.repo.Update(ctx, tournament.FromDomain(t))
	if err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: error updating tournament: %w", op, err)
	}

	return updated.ToDomain(), nil
}

func (uc Usecase) BeginRegistration(ctx context.Context, id int) (tourDomain.Tournament, error) {
	const op = "usecase.tournament.BeginRegistration"

	tour, err := uc.ByID(ctx, id)
	if err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := tour.BeginRegistration(); err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: %w", op, err)
	}

	updated, err := uc.repo.UpdateWithEvent(ctx, tournament.FromDomain(tour))
	if err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: error updating tournament: %w", op, err)
	}

	return updated.ToDomain(), nil
}

func (uc Usecase) Cancel(ctx context.Context, id int) (tourDomain.Tournament, error) {
	const op = "usecase.tournament.Cancel"

	tour, err := uc.ByID(ctx, id)
	if err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := tour.Cancel(); err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: %w", op, err)
	}

	updated, err := uc.repo.UpdateWithEvent(ctx, tournament.FromDomain(tour))
	if err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: error updating tournament: %w", op, err)
	}

	return updated.ToDomain(), nil
}

func (uc Usecase) Start(ctx context.Context, id int) (tourDomain.Tournament, error) {
	const op = "usecase.tournament.Start"

	tour, err := uc.ByID(ctx, id)
	if err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: %w", op, err)
	}

	participants, err := uc.participantRepo.List(ctx, domain.ParticipantFilter{TournamentID: id})
	if err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: error getting participants: %w", op, err)
	}

	pts := make([]domain.Participant, 0, len(participants))
	for _, p := range participants {
		pts = append(pts, p.ToDomain())
	}

	if err := tour.CheckConditions(pts); err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := tour.Start(); err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: %w", op, err)
	}

	updated, err := uc.repo.UpdateWithEvent(ctx, tournament.FromDomain(tour))
	if err != nil {
		return tourDomain.Tournament{}, fmt.Errorf("%s: error updating tournament: %w", op, err)
	}

	return updated.ToDomain(), nil
}
