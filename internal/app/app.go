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
	outboxTournamentWorker "github.com/OndasAlikhan/tourik/internal/outbox/tournament"
	outboxRepo "github.com/OndasAlikhan/tourik/internal/repo/outbox"
	participantRepo "github.com/OndasAlikhan/tourik/internal/repo/participant"
	tournamentRepo "github.com/OndasAlikhan/tourik/internal/repo/tournament"
	transportHttp "github.com/OndasAlikhan/tourik/internal/transport/http"
	participantHandler "github.com/OndasAlikhan/tourik/internal/transport/http/participant"
	tournamentHandler "github.com/OndasAlikhan/tourik/internal/transport/http/tournament"
	participantUc "github.com/OndasAlikhan/tourik/internal/usecase/participant"
	tournamentUc "github.com/OndasAlikhan/tourik/internal/usecase/tournament"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type App struct {
	Logger       *slog.Logger
	Container    internal.Container
	db           *pgxpool.Pool
	server       *http.Server
	config       config.AppConfig
	kafkaWriter  *kafka.Writer
	outboxWorker *outboxTournamentWorker.Worker
	cancel       context.CancelFunc
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

	outboxRepo := outboxRepo.New(sqlxDB)

	tournamentRepo := tournamentRepo.New(sqlxDB, outboxRepo, a.config)
	participantRepo := participantRepo.New(sqlxDB, outboxRepo, a.config)

	tournamentUc := tournamentUc.New(tournamentRepo, participantRepo)
	tournamentHandler := tournamentHandler.New(tournamentUc)

	participantUc := participantUc.New(participantRepo, tournamentUc)
	participantHandler := participantHandler.New(participantUc)

	a.Container = internal.Container{
		Handlers: internal.Handlers{
			TournamentHandler:  tournamentHandler,
			ParticipantHandler: participantHandler,
		},
	}
	a.db = db

	kafkaWriter := NewKafkaWriter(a.config)
	a.kafkaWriter = kafkaWriter

	a.outboxWorker = outboxTournamentWorker.New(kafkaWriter, *a.Logger, outboxTournamentWorker.Config{
		Period:    a.config.OutboxPeriod,
		BatchSize: a.config.OutboxBatchSize,
	}, outboxRepo)

	return nil
}

func (a *App) Run() {
	if err := RunMigrations(a.db); err != nil {
		a.Logger.Error("error running migrations", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	go a.outboxWorker.Run(ctx)

	router := transportHttp.Routers(a.Container)
	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", a.config.Port),
		Handler: router,
	}

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.Logger.Error("error running server", "error", err)
	}
}

func (a *App) Stop(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			a.Logger.Error("error shutting down server", "error", err)
		}
	}
	if a.kafkaWriter != nil {
		if err := a.kafkaWriter.Close(); err != nil {
			a.Logger.Error("error closing kafka writer", "error", err)
		}
	}
	if a.db != nil {
		a.db.Close()
	}
	_ = os.Stderr.Sync()
}
