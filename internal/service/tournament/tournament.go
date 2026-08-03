package tournament

import (
	"context"
	"fmt"

	"github.com/OndasAlikhan/tourik/internal/domain"
	"github.com/OndasAlikhan/tourik/internal/repo/tournament"
)

type Service struct {
	repo tournament.Repo
}

func New(repo tournament.Repo) Service {
	return Service{
		repo: repo,
	}
}

func (s Service) List(ctx context.Context) (result []domain.Tournament, err error) {
	const op = "service.tournament.List"
	tournaments, err := s.repo.List(ctx)
	if err != nil {
		return []domain.Tournament{}, fmt.Errorf("%s: error getting tournament list: %w", op, err)
	}

	for _, val := range tournaments {
		result = append(result, val.ToDomain())
	}

	return result, nil
}
