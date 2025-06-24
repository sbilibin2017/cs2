package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/repositories"
	"github.com/sbilibin2017/cs2/internal/types"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupClickHouseContainer(t *testing.T) (clickhouse.Conn, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "clickhouse/clickhouse-server:21.8.10.19",
		ExposedPorts: []string{"9000/tcp"},
		WaitingFor:   wait.ForListeningPort("9000/tcp"),
	}

	clickhouseC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := clickhouseC.Host(ctx)
	require.NoError(t, err)
	port, err := clickhouseC.MappedPort(ctx, "9000")
	require.NoError(t, err)

	connOpts := &clickhouse.Options{
		Addr: []string{host + ":" + port.Port()},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		DialTimeout: 1 * time.Second,
	}

	conn, err := clickhouse.Open(connOpts)
	require.NoError(t, err)

	createTableSQL := `
CREATE TABLE IF NOT EXISTS games (
    game_id UInt64,
    begin_at DateTime,
    league_id UInt64,
    serie_id UInt64,
    tier_id UInt64,
    tournament_id UInt64,
    map_id UInt64,
    team_id UInt64,
    team_opponent_id UInt64,
    player_id UInt64,
    player_opponent_id UInt64,
    kills UInt32,
    deaths UInt32,
    assists UInt32,
    headshots UInt32,
    flash_assists UInt32,
    kd_diff Float64,
    first_kills_diff Float64,
    adr Float64,
    kast Float64,
    rating Float64,
    round_id UInt64,
    round_outcome_id UInt64,
    round_win UInt64
) ENGINE = Memory
`
	require.NoError(t, conn.Exec(ctx, createTableSQL))

	teardown := func() {
		_ = conn.Close()
		_ = clickhouseC.Terminate(ctx)
	}

	return conn, teardown
}

func TestGameSaverRepository_Save(t *testing.T) {
	ctx := context.Background()
	conn, teardown := setupClickHouseContainer(t)
	defer teardown()

	repo := repositories.NewGameSaverRepository(repositories.WithDB(conn))

	games := []types.GameDB{
		{
			GameID:           1,
			BeginAt:          time.Now(),
			LeagueID:         10,
			SerieID:          20,
			TierID:           30,
			TournamentID:     40,
			MapID:            50,
			TeamID:           70,
			TeamOpponentID:   80,
			PlayerID:         90,
			PlayerOpponentID: 100,
			Kills:            5,
			Deaths:           2,
			Assists:          3,
			Headshots:        1,
			FlashAssists:     1,
			KDDiff:           1.5,
			FirstKillsDiff:   2,
			ADR:              80.5,
			Kast:             75.0,
			Rating:           1.1,
			RoundID:          11,
			RoundOutcomeID:   12,
			RoundWin:         1,
		},
	}

	err := repo.Save(ctx, games)
	require.NoError(t, err)
}
