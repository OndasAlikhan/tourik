package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/OndasAlikhan/tourik/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app := app.NewApp()
	if err := app.Run(ctx); err != nil {
		app.Logger.Error("app.Run", "error", err)
	}

}
