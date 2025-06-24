package apps_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/sbilibin2017/cs2/internal/apps"
	"github.com/sbilibin2017/cs2/internal/configs"
)

func startClickhouseContainer(t *testing.T) (dsn string, terminate func()) {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "clickhouse/clickhouse-server:latest",
			ExposedPorts: []string{"9000/tcp"},
			WaitingFor:   wait.ForListeningPort("9000/tcp").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "9000")
	require.NoError(t, err)

	dsn = fmt.Sprintf("clickhouse://default:@%s:%s/default", host, port.Port())
	return dsn, func() {
		_ = container.Terminate(ctx)
	}
}

func TestApp_Run_WithWorker_Success(t *testing.T) {
	dsn, cleanup := startClickhouseContainer(t)
	defer cleanup()

	cfg := configs.NewConfig(
		configs.WithDatabaseDSN(dsn),
		configs.WithParserDir("./testdata"),
	)

	app, err := apps.NewApp(cfg)
	require.NoError(t, err)
	require.NotNil(t, app)

	// Override workers with dummy successful worker
	app.Workers = []func(ctx context.Context) error{
		func(ctx context.Context) error {
			// simulate short work
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	}

	err = app.Run(context.Background())
	assert.NoError(t, err)
}

func TestApp_Run_WithWorker_ErrorLogged(t *testing.T) {
	dsn, cleanup := startClickhouseContainer(t)
	defer cleanup()

	cfg := configs.NewConfig(
		configs.WithDatabaseDSN(dsn),
		configs.WithParserDir("./testdata"),
	)

	app, err := apps.NewApp(cfg)
	require.NoError(t, err)

	// Inject a worker that returns error
	app.Workers = []func(ctx context.Context) error{
		func(ctx context.Context) error {
			return fmt.Errorf("test error")
		},
	}

	// App should not fail fatally, only log the error
	err = app.Run(context.Background())
	assert.NoError(t, err)
}
