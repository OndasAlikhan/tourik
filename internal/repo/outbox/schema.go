package outbox

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID        int             `db:"id"`
	Topic     string          `db:"topic"`
	Key       string          `db:"key"`
	Message   json.RawMessage `db:"message"`
	CreatedAt time.Time       `db:"created_at"`
	SentAt    *time.Time      `db:"sent_at"`
}
