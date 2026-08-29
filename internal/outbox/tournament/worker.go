package tournament

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/OndasAlikhan/tourik/internal/repo/outbox"
)

type Repo interface {
	ListPending(ctx context.Context, limit int) ([]outbox.Event, error)
	MarkSent(ctx context.Context, id int) error
}

type Worker struct {
	kafka  *kafka.Writer
	repo   Repo
	config Config
	logger slog.Logger
}
type Config struct {
	Period    time.Duration
	BatchSize int
}

func New(kafka *kafka.Writer, logger slog.Logger, config Config, repo Repo) *Worker {
	return &Worker{
		kafka:  kafka,
		repo:   repo,
		config: config,
		logger: logger,
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.logger.Info("tournament outbox worker started")

	ticker := time.NewTicker(w.config.Period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processOutbox(ctx)
		case <-ctx.Done():
			return
		}
	}

}

func (w *Worker) processOutbox(ctx context.Context) {
	events, err := w.repo.ListPending(ctx, w.config.BatchSize)
	if err != nil {
		w.logger.Error("error listing pending outbox events", "error", err)
		return
	}
	if len(events) == 0 {
		return
	}

	messages := make([]kafka.Message, len(events))
	for i, event := range events {
		messages[i] = kafka.Message{
			Topic: event.Topic,
			Key:   []byte(event.Key),
			Value: event.Message,
		}
	}

	if err := w.kafka.WriteMessages(ctx, messages...); err != nil {
		w.logger.Error("error writing outbox events to kafka", "error", err, "batch_size", len(messages))
		return
	}

	for _, event := range events {
		if err := w.repo.MarkSent(ctx, event.ID); err != nil {
			w.logger.Error("error marking outbox event as sent", "error", err, "event_id", event.ID)
		}
	}
}
