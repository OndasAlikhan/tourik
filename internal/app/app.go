package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/OndasAlikhan/tourik/internal/config"
)

type App struct {
	Logger slog.Logger
}

func NewApp() (*App, error) {
	config := config.NewAppConfig()
	_, err := NewDBPool(context.Background(), config)

	if err != nil {
		return nil, fmt.Errorf("NewApp() error creating database: %w", err)
	}

	app := App{}
	return &app, nil
}

func (a *App) Run() error {
	return nil
}

func (a *App) Stop(ctx context.Context) {

}
