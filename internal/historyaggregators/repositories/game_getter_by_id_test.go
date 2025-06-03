package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const createTableQuery2 = `
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

func setupClickHouseContainer2(ctx context.Context, t *testing.T) (clickhouse.Conn, func()) {
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

	err = conn.Exec(ctx, createTableQuery2)
	require.NoError(t, err)

	teardown := func() {
		_ = clickhouseC.Terminate(ctx)
	}

	return conn, teardown
}

func TestGameGetterByIDRepository_Get(t *testing.T) {
	ctx := context.Background()
	conn, teardown := setupClickHouseContainer2(ctx, t)
	defer teardown()

	repo := NewGameGetterByIDRepository(conn)

	// Insert test data
	testUUID := uuid.New()
	now := time.Now().Truncate(time.Second)

	err := conn.Exec(ctx, `
		INSERT INTO games (
			id, begin_at, game_id, league_id, serie_id, tournament_id, tier_id, map_id,
			team_id, team_opponent_id, player_id, player_opponent_id,
			kills, deaths, assists, headshots, flash_assists, first_kills_diff, k_d_diff,
			adr, kast, rating, round_id, round_outcome_id, round_is_ct, round_win, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 1, 1, 1, 1, ?
		)
	`,
		testUUID, now, int32(123), int32(1), int32(2), int32(3), int32(4), int32(5),
		int32(6), int32(7), int32(8), int32(9),
		now,
	)
	require.NoError(t, err)

	// Call method
	games, err := repo.Get(ctx, 123)
	require.NoError(t, err)
	require.NotEmpty(t, games)

	// Check returned data
	found := false
	for _, g := range games {
		if g.ID == testUUID {
			found = true
			assert.Equal(t, int32(123), g.GameID)
			assert.Equal(t, int32(1), g.LeagueID)
			assert.Equal(t, int32(2), g.SerieID)
			assert.Equal(t, int32(3), g.TournamentID)
			assert.Equal(t, int32(4), g.TierID)
			assert.Equal(t, int32(5), g.MapID)
			assert.Equal(t, int32(6), g.TeamID)
			assert.Equal(t, int32(7), g.TeamOpponentID)
			assert.Equal(t, int32(8), g.PlayerID)
			assert.Equal(t, int32(9), g.PlayerOpponentID)
			assert.Equal(t, int32(1), g.RoundID)
			assert.Equal(t, int32(1), g.RoundIsCT)
		}
	}
	assert.True(t, found, "Inserted game must be returned by Get")
}
