package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/sbilibin2017/cs2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const createTableQuery = `
CREATE TABLE IF NOT EXISTS games (
    id UUID,
    begin_at DateTime,

    game_id Int32,
    round_id Int32,
    round_outcome_id Int32,
    round_is_ct Int32,

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

    win Int32,
    updated_at DateTime
) ENGINE = Memory
`

func setupClickHouseContainer(ctx context.Context, t *testing.T) (clickhouse.Conn, func()) {
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

	err = conn.Exec(ctx, createTableQuery)
	require.NoError(t, err)

	teardown := func() {
		_ = clickhouseC.Terminate(ctx)
	}

	return conn, teardown
}

func TestGameSaverRepository_Save(t *testing.T) {
	ctx := context.Background()

	conn, teardown := setupClickHouseContainer(ctx, t)
	defer teardown()

	repo := NewGameSaverRepository(conn)

	gameDBRecords := []types.GameDB{
		{
			ID:               uuid.New(),
			BeginAt:          time.Now().Truncate(time.Second),
			GameID:           1,
			RoundID:          1,
			RoundOutcomeID:   1,
			RoundIsCT:        1,
			LeagueID:         10,
			SerieID:          20,
			TournamentID:     30,
			TierID:           40,
			MapID:            50,
			TeamID:           100,
			TeamOpponentID:   200,
			PlayerID:         300,
			PlayerOpponentID: 400,
			Kills:            5,
			Deaths:           2,
			Assists:          3,
			Headshots:        1,
			FlashAssists:     0,
			FirstKillsDiff:   1,
			KDDiff:           3,
			Adr:              75.5,
			Kast:             65.0,
			Rating:           1.15,
			Win:              1,
			UpdatedAt:        time.Now().Truncate(time.Second),
		},
		{
			ID:               uuid.New(),
			BeginAt:          time.Now().Truncate(time.Second),
			GameID:           2,
			RoundID:          2,
			RoundOutcomeID:   0,
			RoundIsCT:        0,
			LeagueID:         11,
			SerieID:          21,
			TournamentID:     31,
			TierID:           41,
			MapID:            51,
			TeamID:           101,
			TeamOpponentID:   201,
			PlayerID:         301,
			PlayerOpponentID: 401,
			Kills:            10,
			Deaths:           5,
			Assists:          1,
			Headshots:        3,
			FlashAssists:     2,
			FirstKillsDiff:   0,
			KDDiff:           5,
			Adr:              80.2,
			Kast:             70.5,
			Rating:           1.35,
			Win:              0,
			UpdatedAt:        time.Now().Truncate(time.Second),
		},
	}

	err := repo.Save(ctx, gameDBRecords)
	require.NoError(t, err)

	// Query back inserted data to verify
	rows, err := conn.Query(ctx, "SELECT game_id, kills, deaths FROM games ORDER BY game_id")
	require.NoError(t, err)
	defer rows.Close()

	var results []struct {
		GameID int32
		Kills  int32
		Deaths int32
	}
	for rows.Next() {
		var r struct {
			GameID int32
			Kills  int32
			Deaths int32
		}
		err := rows.Scan(&r.GameID, &r.Kills, &r.Deaths)
		require.NoError(t, err)
		results = append(results, r)
	}
	require.NoError(t, rows.Err())

	assert.Len(t, results, 2)
	assert.Equal(t, int32(1), results[0].GameID)
	assert.Equal(t, int32(5), results[0].Kills)
	assert.Equal(t, int32(2), results[0].Deaths)
	assert.Equal(t, int32(2), results[1].GameID)
	assert.Equal(t, int32(10), results[1].Kills)
	assert.Equal(t, int32(5), results[1].Deaths)
}
