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

const playerAggregatorCreateTableQuery = `
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

func setupPlayerAggregatorClickHouseContainer(ctx context.Context, t *testing.T) (clickhouse.Conn, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "clickhouse/clickhouse-server:latest",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			// No password set, default user has no password
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
			Password: "default", // empty password
			Database: "default",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
	})
	require.NoError(t, err)

	err = conn.Exec(ctx, `CREATE DATABASE IF NOT EXISTS default`)
	require.NoError(t, err)

	err = conn.Exec(ctx, playerAggregatorCreateTableQuery)
	require.NoError(t, err)

	teardown := func() {
		_ = clickhouseC.Terminate(ctx)
	}

	return conn, teardown
}

func TestPlayerAggregatorRepository_Aggregate(t *testing.T) {
	ctx := context.Background()
	conn, teardown := setupPlayerAggregatorClickHouseContainer(ctx, t)
	defer teardown()

	repo := NewHistoryAggregatorPlayerRepository(conn)

	now := time.Now()
	gameID := int32(1001)

	insertQuery := `
	INSERT INTO games (
		id, begin_at, game_id, league_id, serie_id, tournament_id, tier_id, map_id,
		team_id, team_opponent_id, player_id, player_opponent_id, kills, deaths, assists,
		headshots, flash_assists, first_kills_diff, k_d_diff, adr, kast, rating,
		round_id, round_outcome_id, round_is_ct, round_win, updated_at
	) VALUES (
		'00000000-0000-0000-0000-000000000000', ?, ?, 1, 1, 1, 1, 1,
		1, 2, 1, 2, 5, 3, 2,
		1, 0, 1, 2, 75.5, 0.6, 1.1,
		?, 1, 1, 1, ?
	)
	`

	// Insert 4 rounds into a single game
	for roundID := 1; roundID <= 4; roundID++ {
		err := conn.Exec(ctx, insertQuery, now.Add(-time.Hour), gameID, roundID, now)
		require.NoError(t, err)
	}

	// Perform aggregation
	agg, err := repo.Aggregate(ctx, now, 1)
	require.NoError(t, err)
	require.NotNil(t, agg)

	// Since aggregation is avg per game, then summed — expect avg per game, not raw total
	expectedGames := float64(1)
	expectedRounds := float64(4)

	// Per-round input: kills=5, deaths=3, assists=2
	// So avg per game = 5, 3, 2
	// And sum of those across 1 game = 5, 3, 2
	expectedKillsTotal := float64(5)
	expectedDeathsTotal := float64(3)
	expectedAssistsTotal := float64(2)

	require.Equal(t, expectedGames, agg.GamesCount)
	require.Equal(t, expectedRounds, agg.RoundsCount)
	require.Equal(t, expectedKillsTotal, agg.KillsTotal)
	require.Equal(t, expectedDeathsTotal, agg.DeathsTotal)
	require.Equal(t, expectedAssistsTotal, agg.AssistsTotal)

	require.InDelta(t, expectedKillsTotal/expectedGames, agg.KillsPerGame, 0.0001)
	require.InDelta(t, expectedKillsTotal/expectedRounds, agg.KillsPerRound, 0.0001)
}
