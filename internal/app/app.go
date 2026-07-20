package app

import (
	"context"
	"log/slog"
)

type App struct {
	Logger slog.Logger
}

func NewApp() *App {
	// config := config.NewAppConfig()

	app := App{}
	return &app
}

func (a *App) Run(ctx context.Context) error {
	<-ctx.Done()
	a.Stop()

	return nil
}

func (a *App) Stop() {

}
