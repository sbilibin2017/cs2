package repositories

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/types"
)

type HistoryAggregtorPlayerRepository struct {
	conn clickhouse.Conn
}

func NewHistoryAggregatorPlayerRepository(conn clickhouse.Conn) *HistoryAggregtorPlayerRepository {
	return &HistoryAggregtorPlayerRepository{
		conn: conn,
	}
}

func (r *HistoryAggregtorPlayerRepository) Aggregate(
	ctx context.Context,
	beginAt time.Time,
	playerID int64,
) (*types.AggregationResult, error) {
	rows, err := r.conn.Query(ctx, historyAggregationQuery, playerID, beginAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var result types.AggregationResult
	err = rows.Scan(
		&result.KillsTotal,
		&result.DeathsTotal,
		&result.AssistsTotal,
		&result.HeadshotsTotal,
		&result.FlashAssistsTotal,

		&result.KillsPerRound,
		&result.DeathsPerRound,
		&result.AssistsPerRound,
		&result.HeadshotsPerRound,
		&result.FlashAssistsPerRound,

		&result.KillsPerGame,
		&result.DeathsPerGame,
		&result.AssistsPerGame,
		&result.HeadshotsPerGame,
		&result.FlashAssistsPerGame,
		&result.FirstKillsDiffPerGame,
		&result.KDdiffPerGame,
		&result.ADRPerGame,
		&result.KASTPerGame,
		&result.RatingPerGame,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

const historyAggregationQuery = `
SELECT
    kills_total,
    deaths_total,
    assists_total,
    headshots_total,
    flash_assists_total,

    kills_per_round,
    deaths_per_round,
    assists_per_round,
    headshots_per_round,
    flash_assists_per_round,

    kills_per_game,
    deaths_per_game,
    assists_per_game,
    headshots_per_game,
    flash_assists_per_game,
    first_kills_diff_per_game,
    kd_diff_per_game,
    adr_per_game,
    kast_per_game,
    rating_per_game
FROM mv_games_player_cumulative
WHERE player_id = ? AND begin_at < ?
ORDER BY begin_at DESC
LIMIT 1
`
