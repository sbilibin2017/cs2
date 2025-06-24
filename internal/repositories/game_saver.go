package repositories

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/types"
)

type GameSaverOption func(*GameSaverRepository)

type GameSaverRepository struct {
	db clickhouse.Conn
}

func WithDB(db clickhouse.Conn) GameSaverOption {
	return func(r *GameSaverRepository) {
		r.db = db
	}
}

func NewGameSaverRepository(opts ...GameSaverOption) *GameSaverRepository {
	repo := &GameSaverRepository{}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func (r *GameSaverRepository) Save(
	ctx context.Context,
	games []types.GameDB,
) error {
	if len(games) == 0 {
		return nil
	}

	batch, err := r.db.PrepareBatch(ctx, saveGamQuery)
	if err != nil {
		return err
	}

	for _, g := range games {
		if err := batch.Append(
			g.GameID,
			g.BeginAt,
			g.LeagueID,
			g.SerieID,
			g.TierID,
			g.TournamentID,
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
			g.KDDiff,
			g.FirstKillsDiff,
			g.ADR,
			g.Kast,
			g.Rating,
			g.RoundID,
			g.RoundOutcomeID,
			g.RoundWin,
		); err != nil {
			return err
		}
	}

	return batch.Send()
}

const saveGamQuery = `
INSERT INTO games (
	game_id, begin_at, 
	league_id, serie_id, tier_id, tournament_id, 
	map_id, 
	team_id, team_opponent_id, player_id, player_opponent_id, 
	kills, deaths, assists, headshots, flash_assists, kd_diff, first_kills_diff, adr, kast, rating, 
	round_id, round_outcome_id, round_win
) VALUES
`
