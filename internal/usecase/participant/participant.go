package participant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/OndasAlikhan/tourik/internal/domain"
	tourDomain "github.com/OndasAlikhan/tourik/internal/domain/tournament"
	"github.com/OndasAlikhan/tourik/internal/repo/participant"
)

type Repo interface {
	ByID(ctx context.Context, id int) (participant.Participant, error)
	List(ctx context.Context, filter domain.ParticipantFilter) ([]participant.Participant, error)
	Create(ctx context.Context, p participant.Participant) (participant.Participant, error)
	Update(ctx context.Context, p participant.Participant) (participant.Participant, error)
	CreateForTournament(ctx context.Context, tournamentID int, p participant.Participant) (participant.Participant, error)
	DeleteFromTournament(ctx context.Context, tournamentID, participantID int) error
}

type TournamentUsecase interface {
	ByID(ctx context.Context, id int) (tourDomain.Tournament, error)
	List(ctx context.Context) (result []tourDomain.Tournament, err error)
}
type Usecase struct {
	repo   Repo
	tourUc TournamentUsecase
}

func New(repo Repo, tourUc TournamentUsecase) Usecase {
	return Usecase{
		repo:   repo,
		tourUc: tourUc,
	}
}

func (uc Usecase) List(ctx context.Context, filter domain.ParticipantFilter) (result []domain.Participant, err error) {
	const op = "usecase.participant.List"
	participants, err := uc.repo.List(ctx, filter)
	if err != nil {
		return []domain.Participant{}, fmt.Errorf("%s: error getting participant list: %w", op, err)
	}

	for _, val := range participants {
		result = append(result, val.ToDomain())
	}

	return result, nil
}

func (uc Usecase) ByID(ctx context.Context, id int) (domain.Participant, error) {
	const op = "usecase.participant.ByID"

	result, err := uc.repo.ByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Participant{}, fmt.Errorf("%s: %w", op, domain.ErrParticipantNotFound)
		}
		return domain.Participant{}, fmt.Errorf("%s: error getting participant: %w", op, err)
	}

	return result.ToDomain(), nil
}

func (uc Usecase) Create(ctx context.Context, p domain.Participant) (domain.Participant, error) {
	const op = "usecase.participant.Create"

	created, err := uc.repo.Create(ctx, participant.FromDomain(p))
	if err != nil {
		return domain.Participant{}, fmt.Errorf("%s: error creating participant: %w", op, err)
	}

	return created.ToDomain(), nil
}

func (uc Usecase) Update(ctx context.Context, p domain.Participant) (domain.Participant, error) {
	const op = "usecase.participant.Update"

	updated, err := uc.repo.Update(ctx, participant.FromDomain(p))
	if err != nil {
		return domain.Participant{}, fmt.Errorf("%s: error updating participant: %w", op, err)
	}

	return updated.ToDomain(), nil
}

func (uc Usecase) CreateForTournament(ctx context.Context, tournamentID int, p domain.Participant) (domain.Participant, error) {
	const op = "usecase.participant.CreateForTournament"

	tour, err := uc.tourUc.ByID(ctx, tournamentID)
	if err != nil {
		return domain.Participant{}, fmt.Errorf("%s: could not get tour: %w", op, err)
	}

	if tour.Status != tourDomain.Registration {
		return domain.Participant{}, fmt.Errorf("%s: tournament not open for registration: %w", op, tourDomain.ErrWrongStatus)
	}

	created, err := uc.repo.CreateForTournament(ctx, tournamentID, participant.FromDomain(p))
	if err != nil {
		return domain.Participant{}, fmt.Errorf("%s: error creating participant for tournament: %w", op, err)
	}

	return created.ToDomain(), nil
}

func (uc Usecase) DeleteFromTournament(ctx context.Context, tournamentID, participantID int) error {
	const op = "usecase.participant.DeleteFromTournament"

	tour, err := uc.tourUc.ByID(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("%s: could not get tour: %w", op, err)
	}

	if tour.Status != tourDomain.Registration {
		return fmt.Errorf("%s: tournament not open for registration: %w", op, tourDomain.ErrWrongStatus)
	}

	if err := uc.repo.DeleteFromTournament(ctx, tournamentID, participantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, domain.ErrParticipantNotFound)
		}
		return fmt.Errorf("%s: error deleting participant from tournament: %w", op, err)
	}

	return nil
}
