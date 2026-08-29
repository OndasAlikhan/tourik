//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	tourDomain "github.com/OndasAlikhan/tourik/internal/domain/tournament"
)

// TestApp_CreateTournament_PublishesToKafka is a smoke test exercising the
// full stack (HTTP -> usecase -> repo -> outbox -> worker -> kafka) against
// real Postgres and Kafka containers.
func TestApp_CreateTournament_PublishesToKafka(t *testing.T) {
	env := newTestEnv(t, 300*time.Millisecond)
	ctx := context.Background()
	env.startWorker()

	id := createTournament(t, env, "Smoke Test Open")
	key := strconv.Itoa(id)

	waitForOutboxSentCount(t, ctx, env, []string{key}, 1, 15*time.Second)

	brokers, err := env.kafka.Brokers(ctx)
	if err != nil {
		t.Fatalf("get kafka brokers: %v", err)
	}

	messages := consumeMessages(t, ctx, brokers, env.cfg.OutboxTopic, 1, 15*time.Second)
	if got := string(messages[0].Key); got != key {
		t.Fatalf("expected message key %q, got %q", key, got)
	}

	var evt tourDomain.TournamentEventMessage
	if err := json.Unmarshal(messages[0].Value, &evt); err != nil {
		t.Fatalf("unmarshal event message: %v", err)
	}
	if evt.EventType != tourDomain.TournamentCreated {
		t.Fatalf("expected event type %q, got %q", tourDomain.TournamentCreated, evt.EventType)
	}
}

// TestOutboxWorker_DeliversMessagesAfterKafkaRecovers verifies the outbox
// pattern's core guarantee: while Kafka is unreachable, tournament events
// created via the API accumulate as pending rows in the outbox table instead
// of being lost or blocking writes. Once Kafka comes back online, the
// background worker must drain and deliver every pending event without any
// additional intervention.
func TestOutboxWorker_DeliversMessagesAfterKafkaRecovers(t *testing.T) {
	env := newTestEnv(t, 300*time.Millisecond)
	ctx := context.Background()

	env.startWorker()

	env.stopKafka(ctx)

	const numTournaments = 3
	ids := make([]int, numTournaments)
	keys := make([]string, numTournaments)
	for i := range ids {
		ids[i] = createTournament(t, env, fmt.Sprintf("Offline Cup #%d", i+1))
		keys[i] = strconv.Itoa(ids[i])
	}

	// Give the worker a few ticks to try (and fail) delivering while Kafka is
	// down, then confirm nothing was lost or incorrectly marked as sent.
	time.Sleep(1200 * time.Millisecond)

	pending, err := countPendingOutboxEvents(ctx, env, keys)
	if err != nil {
		t.Fatalf("count pending outbox events: %v", err)
	}
	if pending != numTournaments {
		t.Fatalf("expected %d pending outbox events while kafka is down, got %d", numTournaments, pending)
	}
	sent, err := countSentOutboxEvents(ctx, env, keys)
	if err != nil {
		t.Fatalf("count sent outbox events: %v", err)
	}
	if sent != 0 {
		t.Fatalf("expected 0 sent outbox events while kafka is down, got %d", sent)
	}

	env.startKafka(ctx)

	waitForOutboxSentCount(t, ctx, env, keys, numTournaments, 30*time.Second)

	brokers, err := env.kafka.Brokers(ctx)
	if err != nil {
		t.Fatalf("get kafka brokers: %v", err)
	}

	messages := consumeMessages(t, ctx, brokers, env.cfg.OutboxTopic, numTournaments, 30*time.Second)
	if err := containsAll(messageKeys(messages), keys); err != nil {
		t.Fatalf("kafka did not receive all outbox messages: %v", err)
	}
}
