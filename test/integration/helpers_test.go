//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	kafkago "github.com/segmentio/kafka-go"
)

// TestMain switches the working directory to the repo root before any tests
// run, since app.RunMigrations resolves the "migrations" folder relative to
// the process's current directory (matching how the production binary is
// launched from the repo root).
func TestMain(m *testing.M) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not determine repo root: runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	if err := os.Chdir(repoRoot); err != nil {
		panic(fmt.Sprintf("chdir to repo root %q: %v", repoRoot, err))
	}

	os.Exit(m.Run())
}

func newTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type createTournamentBody struct {
	Name            string    `json:"Name"`
	Discipline      string    `json:"Discipline"`
	BracketFormat   string    `json:"BracketFormat"`
	MaxParticipants int       `json:"MaxParticipants"`
	StartDate       time.Time `json:"StartDate"`
	EndDate         time.Time `json:"EndDate"`
}

type apiResponse struct {
	Error string `json:"error_message"`
	Data  struct {
		ID int `json:"ID"`
	} `json:"data"`
}

// createTournament posts a new tournament through the real HTTP API and
// returns its generated ID.
func createTournament(t *testing.T, env *testEnv, name string) int {
	t.Helper()

	body := createTournamentBody{
		Name:            name,
		Discipline:      "chess",
		BracketFormat:   "SingleElimination",
		MaxParticipants: 4,
		StartDate:       time.Now().Add(24 * time.Hour),
		EndDate:         time.Now().Add(48 * time.Hour),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal create tournament request: %v", err)
	}

	resp, err := http.Post(env.baseURL()+"/api/tournaments", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create tournament request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read create tournament response: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create tournament: expected status 201, got %d, body: %s", resp.StatusCode, respBody)
	}

	var decoded apiResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		t.Fatalf("decode create tournament response: %v, body: %s", err, respBody)
	}
	if decoded.Data.ID == 0 {
		t.Fatalf("create tournament: expected non-zero id, body: %s", respBody)
	}

	return decoded.Data.ID
}

// waitForOutboxSentCount polls the outbox table until exactly wantSent of the
// given tournament keys have sent_at set, or fails the test on timeout.
func waitForOutboxSentCount(t *testing.T, ctx context.Context, env *testEnv, keys []string, wantSent int, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var lastSent int
	for time.Now().Before(deadline) {
		sent, err := countSentOutboxEvents(ctx, env, keys)
		if err != nil {
			t.Fatalf("count sent outbox events: %v", err)
		}
		lastSent = sent
		if sent == wantSent {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d outbox events to be sent, got %d", wantSent, lastSent)
}

func countSentOutboxEvents(ctx context.Context, env *testEnv, keys []string) (int, error) {
	return countOutboxEvents(ctx, env, keys, `select count(*) from outbox where key in (?) and sent_at is not null`)
}

func countPendingOutboxEvents(ctx context.Context, env *testEnv, keys []string) (int, error) {
	return countOutboxEvents(ctx, env, keys, `select count(*) from outbox where key in (?) and sent_at is null`)
}

func countOutboxEvents(ctx context.Context, env *testEnv, keys []string, baseQuery string) (int, error) {
	query, args, err := sqlx.In(baseQuery, keys)
	if err != nil {
		return 0, err
	}
	query = env.sqlxDB.Rebind(query)

	var count int
	if err := env.sqlxDB.GetContext(ctx, &count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

// consumeMessages reads exactly wantCount messages from the given topic
// starting at the earliest offset, failing the test on timeout.
func consumeMessages(t *testing.T, ctx context.Context, brokers []string, topic string, wantCount int, timeout time.Duration) []kafkago.Message {
	t.Helper()

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  "",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	reader.SetOffset(kafkago.FirstOffset)
	defer reader.Close()

	readCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	messages := make([]kafkago.Message, 0, wantCount)
	for len(messages) < wantCount {
		msg, err := reader.ReadMessage(readCtx)
		if err != nil {
			t.Fatalf("consume kafka messages: got %d/%d, error: %v", len(messages), wantCount, err)
		}
		messages = append(messages, msg)
	}
	return messages
}

func messageKeys(messages []kafkago.Message) []string {
	keys := make([]string, len(messages))
	for i, m := range messages {
		keys[i] = string(m.Key)
	}
	return keys
}

func containsAll(haystack []string, needles []string) error {
	set := make(map[string]struct{}, len(haystack))
	for _, h := range haystack {
		set[h] = struct{}{}
	}
	for _, n := range needles {
		if _, ok := set[n]; !ok {
			return fmt.Errorf("missing key %q in %v", n, haystack)
		}
	}
	return nil
}
