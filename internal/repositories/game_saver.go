package repositories

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

type GameSaverRepository struct {
	db clickhouse.Conn
}

func NewGameSaverRepository(db clickhouse.Conn) *GameSaverRepository {
	return &GameSaverRepository{db: db}
}

func (r *GameSaverRepository) Save(ctx context.Context, games []types.GameDB) error {
	batch, err := r.db.PrepareBatch(ctx, gameSaveQuery)
	if err != nil {
		logger.Log.Error(err)
		return err
	}

	for _, g := range games {
		err := batch.Append(
			g.ID,
			g.BeginAt,
			g.GameID,
			g.RoundID,
			g.RoundOutcomeID,
			g.RoundIsCT,
			g.LeagueID,
			g.SerieID,
			g.TournamentID,
			g.TierID,
			g.MapID,
			g.TeamID,
			g.TeamOpponentID,
			g.PlayerID,
			g.PlayerOpponentID,
			g.Kills,
			g.Deaths,
			g.Assists,
			g.Headshots,
			g.FlashAssists,
			g.FirstKillsDiff,
			g.KDDiff,
			g.Adr,
			g.Kast,
			g.Rating,
			g.Win,
			g.UpdatedAt,
		)
		if err != nil {
			logger.Log.Error(err)
			return err
		}
	}

	if err := batch.Send(); err != nil {
		logger.Log.Error(err)
		return err
	}

	return nil
}

const gameSaveQuery = `
	INSERT INTO games (
		id,
		begin_at,
		game_id,
		round_id,
		round_outcome_id,
		round_is_ct,
		league_id,
		serie_id,
		tournament_id,
		tier_id,
		map_id,
		team_id,
		team_opponent_id,
		player_id,
		player_opponent_id,
		kills,
		deaths,
		assists,
		headshots,
		flash_assists,
		first_kills_diff,
		k_d_diff,
		adr,
		kast,
		rating,
		win,
		updated_at
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	)
`
