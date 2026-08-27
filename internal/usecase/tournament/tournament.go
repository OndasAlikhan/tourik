package tournament

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/OndasAlikhan/tourik/internal/domain"
	"github.com/OndasAlikhan/tourik/internal/repo/tournament"
)

type Repo interface {
	ByID(ctx context.Context, id int) (tournament.Tournament, error)
	List(ctx context.Context) ([]tournament.Tournament, error)
	Create(ctx context.Context, t tournament.Tournament) (tournament.Tournament, error)
	Update(ctx context.Context, t tournament.Tournament) (tournament.Tournament, error)
}
type Usecase struct {
	repo Repo
}

func New(repo tournament.Repo) Usecase {
	return Usecase{
		repo: repo,
	}
}

func (uc Usecase) List(ctx context.Context) (result []domain.Tournament, err error) {
	const op = "usecase.tournament.List"
	tournaments, err := uc.repo.List(ctx)
	if err != nil {
		return []domain.Tournament{}, fmt.Errorf("%s: error getting tournament list: %w", op, err)
	}

	for _, val := range tournaments {
		result = append(result, val.ToDomain())
	}

	return result, nil
}

func (uc Usecase) ByID(ctx context.Context, id int) (domain.Tournament, error) {
	const op = "usecase.tournament.ByID"

	result, err := uc.repo.ByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Tournament{}, fmt.Errorf("%s: %w", op, domain.ErrTourNotFound)
		}
		return domain.Tournament{}, fmt.Errorf("%s: error getting tournament: %w", op, err)
	}

	return result.ToDomain(), nil
}

func (uc Usecase) Create(ctx context.Context, t domain.Tournament) (domain.Tournament, error) {
	const op = "usecase.tournament.Create"

	created, err := uc.repo.Create(ctx, tournament.FromDomain(t))
	if err != nil {
		return domain.Tournament{}, fmt.Errorf("%s: error creating tournament: %w", op, err)
	}

	return created.ToDomain(), nil
}

func (uc Usecase) Update(ctx context.Context, t domain.Tournament) (domain.Tournament, error) {
	const op = "usecase.tournament.Update"

	updated, err := uc.repo.Update(ctx, tournament.FromDomain(t))
	if err != nil {
		return domain.Tournament{}, fmt.Errorf("%s: error updating tournament: %w", op, err)
	}

	return updated.ToDomain(), nil
}
