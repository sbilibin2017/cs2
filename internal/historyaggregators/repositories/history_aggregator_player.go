package repositories

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/historyaggregators/types"
)

type HistoryAggregatorPlayerRepository struct {
	db clickhouse.Conn
}

func NewHistoryAggregatorPlayerRepository(db clickhouse.Conn) *HistoryAggregatorPlayerRepository {
	return &HistoryAggregatorPlayerRepository{db: db}
}

func (r *HistoryAggregatorPlayerRepository) Aggregate(
	ctx context.Context,
	beginAt time.Time,
	playerID int32,
) (*types.Aggregation, error) {
	row := r.db.QueryRow(ctx, playerAggregatorQuery, playerID, beginAt)

	var agg types.Aggregation
	var gamesCount float64
	var roundsCount float64

	err := row.Scan(
		&agg.KillsTotal,
		&agg.DeathsTotal,
		&agg.AssistsTotal,
		&agg.HeadshotsTotal,
		&agg.FlashAssistsTotal,
		&agg.FirstKillsDiffTotal,
		&agg.KDDiffTotal,
		&agg.AdrTotal,
		&agg.KastTotal,
		&agg.RatingTotal,

		&agg.WinRoundTotal,
		&agg.WinRoundH1Total,
		&agg.WinRoundH2Total,
		&agg.IsCTRoundTotal,

		&gamesCount,
		&roundsCount,
	)
	if err != nil {
		return nil, err
	}

	agg.GamesCount = gamesCount
	agg.RoundsCount = roundsCount

	if gamesCount == 0 {
		gamesCount = 1
	}

	agg.KillsPerGame = agg.KillsTotal / float64(gamesCount)
	agg.DeathsPerGame = agg.DeathsTotal / float64(gamesCount)
	agg.AssistsPerGame = agg.AssistsTotal / float64(gamesCount)
	agg.HeadshotsPerGame = agg.HeadshotsTotal / float64(gamesCount)
	agg.FlashAssistsPerGame = agg.FlashAssistsTotal / float64(gamesCount)
	agg.FirstKillsDiffPerGame = agg.FirstKillsDiffTotal / float64(gamesCount)
	agg.KDDiffPerGame = agg.KDDiffTotal / float64(gamesCount)
	agg.AdrPerGame = agg.AdrTotal / float64(gamesCount)
	agg.KastPerGame = agg.KastTotal / float64(gamesCount)
	agg.RatingPerGame = agg.RatingTotal / float64(gamesCount)

	agg.WinRoundPerGame = agg.WinRoundTotal / float64(gamesCount)
	agg.WinRoundH1PerGame = agg.WinRoundH1Total / float64(gamesCount)
	agg.WinRoundH2PerGame = agg.WinRoundH2Total / float64(gamesCount)
	agg.IsCTRoundPerGame = agg.IsCTRoundTotal / float64(gamesCount)

	if roundsCount == 0 {
		roundsCount = 1
	}

	agg.KillsPerRound = agg.KillsTotal / float64(roundsCount)
	agg.DeathsPerRound = agg.DeathsTotal / float64(roundsCount)
	agg.AssistsPerRound = agg.AssistsTotal / float64(roundsCount)
	agg.HeadshotsPerRound = agg.HeadshotsTotal / float64(roundsCount)
	agg.FlashAssistsPerRound = agg.FlashAssistsTotal / float64(roundsCount)
	agg.FirstKillsDiffPerRound = agg.FirstKillsDiffTotal / float64(roundsCount)
	agg.KDDiffPerRound = agg.KDDiffTotal / float64(roundsCount)
	agg.AdrPerRound = agg.AdrTotal / float64(roundsCount)
	agg.KastPerRound = agg.KastTotal / float64(roundsCount)
	agg.RatingPerRound = agg.RatingTotal / float64(roundsCount)

	return &agg, nil
}

const playerAggregatorQuery = `
WITH max_rounds AS (
    SELECT game_id, max(round_id) AS max_round_id
    FROM games
    GROUP BY game_id
)

SELECT
    toFloat64(sum(coalesce(avg_kills, 0))) AS kills_total,
    toFloat64(sum(coalesce(avg_deaths, 0))) AS deaths_total,
    toFloat64(sum(coalesce(avg_assists, 0))) AS assists_total,
    toFloat64(sum(coalesce(avg_headshots, 0))) AS headshots_total,
    toFloat64(sum(coalesce(avg_flash_assists, 0))) AS flash_assists_total,
    toFloat64(sum(coalesce(avg_first_kills_diff, 0))) AS first_kills_diff_total,
    toFloat64(sum(coalesce(avg_k_d_diff, 0))) AS k_d_diff_total,
    toFloat64(sum(coalesce(avg_adr, 0))) AS adr_total,
    toFloat64(sum(coalesce(avg_kast, 0))) AS kast_total,
    toFloat64(sum(coalesce(avg_rating, 0))) AS rating_total,

    toFloat64(sum(coalesce(avg_win_round, 0))) AS win_round_total,
    toFloat64(sum(coalesce(avg_win_round_h1, 0))) AS win_round_h1_total,
    toFloat64(sum(coalesce(avg_win_round_h2, 0))) AS win_round_h2_total,
    toFloat64(sum(coalesce(avg_is_ct_round, 0))) AS is_ct_round_total,

    toFloat64(count(*)) AS games_count,
    toFloat64(sum(coalesce(rounds_count, 0))) AS rounds_count

FROM (
    SELECT
        game_id,
        avg(kills) AS avg_kills,
        avg(deaths) AS avg_deaths,
        avg(assists) AS avg_assists,
        avg(headshots) AS avg_headshots,
        avg(flash_assists) AS avg_flash_assists,
        avg(first_kills_diff) AS avg_first_kills_diff,
        avg(k_d_diff) AS avg_k_d_diff,
        avg(adr) AS avg_adr,
        avg(kast) AS avg_kast,
        avg(rating) AS avg_rating,

        avg(round_win) AS avg_win_round,
        avg(round_win_h1) AS avg_win_round_h1,
        avg(round_win_h2) AS avg_win_round_h2,
        avg(round_is_ct) AS avg_is_ct_round,

        count(round_id) AS rounds_count

    FROM (
        SELECT
            g.game_id,
            g.kills,
            g.deaths,
            g.assists,
            g.headshots,
            g.flash_assists,
            g.first_kills_diff,
            g.k_d_diff,
            g.adr,
            g.kast,
            g.rating,
            g.round_win,
            g.round_id,
            g.round_is_ct,

            CASE WHEN g.round_id <= max_rounds.max_round_id / 2 THEN g.round_win ELSE 0 END AS round_win_h1,
            CASE WHEN g.round_id > max_rounds.max_round_id / 2 THEN g.round_win ELSE 0 END AS round_win_h2

        FROM games g
        INNER JOIN max_rounds ON g.game_id = max_rounds.game_id
        WHERE g.player_id = ? AND g.begin_at < ?
    ) subgames
    GROUP BY game_id
) agg;
`
