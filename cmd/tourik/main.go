package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	appFactory "github.com/OndasAlikhan/tourik/internal/app"
)

func main() {
	app, err := appFactory.NewApp()
	if err != nil {
		slog.Error("app.NewApp error:", "error", err)
		return
	}

	app.Logger.Info("starting app")
	go app.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel() // это стоит вызвать внутри app.Stop() ?
	app.Logger.Info("stopping app")
	app.Stop(ctx)

}
