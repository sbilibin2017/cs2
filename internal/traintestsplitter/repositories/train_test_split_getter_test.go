package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"github.com/testcontainers/testcontainers-go/wait"
)

const createTableQuery = `
CREATE TABLE IF NOT EXISTS games (
	id UUID,
	begin_at DateTime,
	game_id Int32,
	league_id Int32,
	serie_id Int32,
	tournament_id Int32,
	tier_id Int32,
	map_id Int32,
	team_id Int32,
	team_opponent_id Int32,
	player_id Int32,
	player_opponent_id Int32,
	kills Int32,
	deaths Int32,
	assists Int32,
	headshots Int32,
	flash_assists Int32,
	first_kills_diff Int32,
	k_d_diff Int32,
	adr Float32,
	kast Float32,
	rating Float32,
	round_id Int32,
	round_outcome_id Int32,
	round_is_ct Int32,
	round_win Int32,
	updated_at DateTime
) ENGINE = Memory
`

func setupClickHouseContainer(ctx context.Context, t *testing.T) (clickhouse.Conn, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "clickhouse/clickhouse-server:latest",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"CLICKHOUSE_USER":     "default",
			"CLICKHOUSE_PASSWORD": "default",
			"CLICKHOUSE_DATABASE": "default",
		},
		WaitingFor: wait.ForListeningPort("9000/tcp").WithStartupTimeout(60 * time.Second),
	}

	clickhouseC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := clickhouseC.Host(ctx)
	require.NoError(t, err)

	port, err := clickhouseC.MappedPort(ctx, nat.Port("9000"))
	require.NoError(t, err)

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{host + ":" + port.Port()},
		Auth: clickhouse.Auth{
			Username: "default",
			Password: "default",
			Database: "default",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
	})
	require.NoError(t, err)

	err = conn.Exec(ctx, `CREATE DATABASE IF NOT EXISTS default`)
	require.NoError(t, err)

	err = conn.Exec(ctx, createTableQuery)
	require.NoError(t, err)

	teardown := func() {
		_ = clickhouseC.Terminate(ctx)
	}

	return conn, teardown
}

func TestTrainTestSplitGetterRepository_Get(t *testing.T) {
	ctx := context.Background()

	conn, teardown := setupClickHouseContainer(ctx, t)
	defer teardown()

	repo := NewTrainTestSplitGetterRepository(conn)

	// Insert test data
	for i := 1; i <= 150; i++ {
		err := conn.Exec(ctx, `
			INSERT INTO games (id, begin_at, game_id, updated_at)
			VALUES (?, ?, ?, ?)`,
			// Use dummy UUID, example: all zeros, or generate random ones
			"00000000-0000-0000-0000-000000000000",
			time.Now().Add(time.Duration(i)*time.Minute),
			i,
			time.Now(),
		)
		require.NoError(t, err)
	}

	// Run the query to get train/test split
	split, err := repo.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, split)

	// Assert test split length = 100 (last 100 game_ids)
	require.Len(t, split.TestGameIDs, 100)

	// Assert train split length = 50 (rest)
	require.Len(t, split.TrainGameIDs, 50)

	// The highest game_id should be in test set (because we sorted descending)
	require.Contains(t, split.TestGameIDs, int32(150))

	// The lowest game_id should be in train set
	require.Contains(t, split.TrainGameIDs, int32(1))
}
