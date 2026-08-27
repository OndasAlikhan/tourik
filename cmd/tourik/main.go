package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/OndasAlikhan/tourik/docs"
	appFactory "github.com/OndasAlikhan/tourik/internal/app"
)

const stopTimeout = time.Second * 10

func main() {
	app, err := appFactory.NewApp()
	if err != nil {
		slog.Error("app.NewApp error:", "error", err)
		return
	}

	err = app.InitContainer()
	if err != nil {
		slog.Error("failed to init container", "error", err)
		return
	}

	app.Logger.Info("starting app")
	go app.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("stopping app")
	ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()
	app.Stop(ctx)

}
