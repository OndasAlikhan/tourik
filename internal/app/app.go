package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/OndasAlikhan/tourik/internal"
	"github.com/OndasAlikhan/tourik/internal/config"
	tournamentRepo "github.com/OndasAlikhan/tourik/internal/repo/tournament"
	tournamentService "github.com/OndasAlikhan/tourik/internal/service/tournament"
	"github.com/OndasAlikhan/tourik/internal/transport/http"
	tournamentHandler "github.com/OndasAlikhan/tourik/internal/transport/http/tournament"
)

type App struct {
	Logger    slog.Logger
	Container internal.Container
}

func NewApp() (*App, error) {
	conf := config.NewAppConfig()
	db, err := NewDBPool(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("NewApp() error creating database: %w", err)
	}

	tournamentRepo := tournamentRepo.New(db)
	tournamentService := tournamentService.New(tournamentRepo)
	tournamentHandler := tournamentHandler.New(tournamentService)

	container := internal.Container{
		Handlers: internal.Handlers{
			TournamentHandler: tournamentHandler,
		},
	}

	app := App{
		Container: container,
	}
	return &app, nil
}

func (a *App) Run() {
	router := http.Routers(a.Container)
	err := router.Run()
	if err != nil {
		a.Logger.Error("error running router", "error", err)
		return
	}
}

func (a *App) Stop(ctx context.Context) {

}
