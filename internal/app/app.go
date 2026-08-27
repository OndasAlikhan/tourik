package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/OndasAlikhan/tourik/internal"
	"github.com/OndasAlikhan/tourik/internal/config"
	tournamentRepo "github.com/OndasAlikhan/tourik/internal/repo/tournament"
	transportHttp "github.com/OndasAlikhan/tourik/internal/transport/http"
	tournamentHandler "github.com/OndasAlikhan/tourik/internal/transport/http/tournament"
	tournamentService "github.com/OndasAlikhan/tourik/internal/usecase/tournament"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Logger    *slog.Logger
	Container internal.Container
	db        *pgxpool.Pool
	server    *http.Server
	config    config.AppConfig
}

func NewApp() (*App, error) {
	return &App{}, nil
}

func (a *App) InitContainer() error {
	a.Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

	a.config = config.NewAppConfig()
	db, err := NewDBPool(context.Background(), a.config)
	if err != nil {
		return fmt.Errorf("NewApp() error creating database: %w", err)
	}

	sqlxDB := NewSqlxDB(db)

	tournamentRepo := tournamentRepo.New(sqlxDB)
	tournamentService := tournamentService.New(tournamentRepo)
	tournamentHandler := tournamentHandler.New(tournamentService)

	a.Container = internal.Container{
		Handlers: internal.Handlers{
			TournamentHandler: tournamentHandler,
		},
	}
	a.db = db

	return nil
}

func (a *App) Run() {
	if err := RunMigrations(a.db); err != nil {
		a.Logger.Error("error running migrations", "error", err)
		return
	}

	router := transportHttp.Routers(a.Container)
	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.Port),
		Handler: router,
	}

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.Logger.Error("error running server", "error", err)
	}
}

func (a *App) Stop(ctx context.Context) {
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			a.Logger.Error("error shutting down server", "error", err)
		}
	}
	if a.db != nil {
		a.db.Close()
	}
	_ = os.Stderr.Sync()
}
