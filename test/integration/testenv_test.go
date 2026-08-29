//go:build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	kafkago "github.com/segmentio/kafka-go"
	tc "github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/OndasAlikhan/tourik/internal"
	apppkg "github.com/OndasAlikhan/tourik/internal/app"
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
)

// testEnv wires up the same components as app.App (db, kafka writer, outbox
// worker, http handlers) but pointed at ephemeral testcontainers instances so
// that the wiring under test matches production as closely as possible.
type testEnv struct {
	t *testing.T

	pg    *postgres.PostgresContainer
	kafka *tckafka.KafkaContainer

	cfg config.AppConfig

	db     *pgxpool.Pool
	sqlxDB *sqlx.DB

	kafkaWriter *kafkago.Writer
	outbox      outboxRepo.Repo
	worker      *outboxTournamentWorker.Worker

	server *httptest.Server
}

func newTestEnv(t *testing.T, outboxPeriod time.Duration) *testEnv {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("tourik"),
		postgres.WithUsername("tourik"),
		postgres.WithPassword("tourik"),
		tc.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := tc.TerminateContainer(pgContainer); err != nil {
			t.Logf("terminate postgres container: %v", err)
		}
	})

	// Docker (re)assigns a fresh ephemeral host port every time a container
	// with an unspecified ("0") host port is started, including on a plain
	// restart of the same container. Pin the broker to a fixed host port so
	// its address stays stable across the stop/start cycle the kafka-outage
	// test performs.
	kafkaHostPort := pickFreeHostPort(t)
	kafkaContainer, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.6.1",
		withFixedHostPort("9093/tcp", kafkaHostPort),
	)
	if err != nil {
		t.Fatalf("start kafka container: %v", err)
	}
	t.Cleanup(func() {
		if err := tc.TerminateContainer(kafkaContainer); err != nil {
			t.Logf("terminate kafka container: %v", err)
		}
	})

	pgHost, err := pgContainer.Host(ctx)
	if err != nil {
		t.Fatalf("postgres host: %v", err)
	}
	pgPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("postgres mapped port: %v", err)
	}

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatalf("kafka brokers: %v", err)
	}

	cfg := config.AppConfig{
		DatabaseName:     "tourik",
		DatabaseHost:     pgHost,
		DatabasePort:     pgPort.Port(),
		DatabaseUser:     "tourik",
		DatabasePassword: "tourik",
		KafkaBroker:      brokers[0],
		OutboxTopic:      "tournament-events",
		OutboxPeriod:     outboxPeriod,
		OutboxBatchSize:  100,
	}

	db, err := apppkg.NewDBPool(ctx, cfg)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	t.Cleanup(db.Close)

	if err := apppkg.RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	sqlxDB := apppkg.NewSqlxDB(db)

	oRepo := outboxRepo.New(sqlxDB)
	tRepo := tournamentRepo.New(sqlxDB, oRepo, cfg)
	pRepo := participantRepo.New(sqlxDB, oRepo, cfg)

	tUc := tournamentUc.New(tRepo, pRepo)
	tHandler := tournamentHandler.New(tUc)

	pUc := participantUc.New(pRepo, tUc)
	pHandler := participantHandler.New(pUc)

	container := internal.Container{
		Handlers: internal.Handlers{
			TournamentHandler:  tHandler,
			ParticipantHandler: pHandler,
		},
	}

	router := transportHttp.Routers(container)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	kafkaWriter := apppkg.NewKafkaWriter(cfg)
	t.Cleanup(func() {
		_ = kafkaWriter.Close()
	})

	logger := *newTestLogger(t)
	worker := outboxTournamentWorker.New(kafkaWriter, logger, outboxTournamentWorker.Config{
		Period:    cfg.OutboxPeriod,
		BatchSize: cfg.OutboxBatchSize,
	}, oRepo)

	return &testEnv{
		t:           t,
		pg:          pgContainer,
		kafka:       kafkaContainer,
		cfg:         cfg,
		db:          db,
		sqlxDB:      sqlxDB,
		kafkaWriter: kafkaWriter,
		outbox:      oRepo,
		worker:      worker,
		server:      server,
	}
}

// startWorker runs the outbox worker in the background until the test ends.
func (e *testEnv) startWorker() {
	ctx, cancel := context.WithCancel(context.Background())
	go e.worker.Run(ctx)
	e.t.Cleanup(cancel)
}

// stopKafka simulates a kafka outage by stopping (not removing) the
// container, so it can be restarted with the same broker address/port.
func (e *testEnv) stopKafka(ctx context.Context) {
	e.t.Helper()
	if err := e.kafka.Stop(ctx, nil); err != nil {
		e.t.Fatalf("stop kafka container: %v", err)
	}
}

// startKafka brings the previously-stopped kafka container back online.
//
// The module's built-in readiness hook waits for a log line that may already
// be present from the container's first boot, so on a restart it can report
// "ready" before the broker is actually accepting connections again. To get
// a trustworthy signal we additionally poll the broker directly until it
// responds to an ApiVersions request.
func (e *testEnv) startKafka(ctx context.Context) {
	e.t.Helper()
	if err := e.kafka.Start(ctx); err != nil {
		e.t.Fatalf("start kafka container: %v", err)
	}

	brokers, err := e.kafka.Brokers(ctx)
	if err != nil {
		e.t.Fatalf("get kafka brokers after restart: %v", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	var lastErr error
	for {
		select {
		case <-waitCtx.Done():
			e.t.Fatalf("kafka broker did not become reachable after restart: %v", lastErr)
		default:
		}

		conn, err := kafkago.DialContext(waitCtx, "tcp", brokers[0])
		if err != nil {
			lastErr = err
			time.Sleep(300 * time.Millisecond)
			continue
		}
		_, err = conn.ApiVersions()
		conn.Close()
		if err != nil {
			lastErr = err
			time.Sleep(300 * time.Millisecond)
			continue
		}
		return
	}
}

// pickFreeHostPort finds a currently-unused TCP port on the host. There is
// an inherent race between releasing it here and Docker binding it, but in
// practice it's reliable enough for test setup.
func pickFreeHostPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free host port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// withFixedHostPort binds containerPort (e.g. "9093/tcp") to a specific host
// port instead of letting Docker allocate a random one, so the mapping
// survives a container stop/start cycle.
func withFixedHostPort(containerPort string, hostPort int) tc.CustomizeRequestOption {
	return tc.WithHostConfigModifier(func(hc *container.HostConfig) {
		port, err := network.ParsePort(containerPort)
		if err != nil {
			panic(fmt.Sprintf("parse container port %q: %v", containerPort, err))
		}
		if hc.PortBindings == nil {
			hc.PortBindings = network.PortMap{}
		}
		hc.PortBindings[port] = []network.PortBinding{{
			HostIP:   netip.IPv4Unspecified(),
			HostPort: strconv.Itoa(hostPort),
		}}
	})
}

func (e *testEnv) baseURL() string {
	return e.server.URL
}
