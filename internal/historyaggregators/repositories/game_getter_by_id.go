package repositories

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/historyaggregators/types"
	"github.com/sbilibin2017/cs2/internal/logger"
)

type GameGetterByIDRepository struct {
	db clickhouse.Conn
}

func NewGameGetterByIDRepository(db clickhouse.Conn) *GameGetterByIDRepository {
	return &GameGetterByIDRepository{db: db}
}

func (r *GameGetterByIDRepository) Get(ctx context.Context, gameID int32) ([]types.GameDB, error) {
	rows, err := r.db.Query(ctx, gameGetterByIDQuery, gameID)
	if err != nil {
		logger.Log.Error(err)
		return nil, err
	}
	defer rows.Close()

	var games []types.GameDB
	for rows.Next() {
		var g types.GameDB
		if err := rows.Scan(
			&g.ID,
			&g.BeginAt,
			&g.GameID,
			&g.LeagueID,
			&g.SerieID,
			&g.TournamentID,
			&g.TierID,
			&g.MapID,
			&g.TeamID,
			&g.TeamOpponentID,
			&g.PlayerID,
			&g.PlayerOpponentID,
			&g.RoundID,
			&g.RoundIsCT,
		); err != nil {
			logger.Log.Error(err)
			return nil, err
		}
		games = append(games, g)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Error(err)
		return nil, err
	}

	return games, nil
}

const gameGetterByIDQuery = `
SELECT
	id,
	begin_at,
	game_id,
	league_id,
	serie_id,
	tournament_id,
	tier_id,
	map_id,
	team_id,
	team_opponent_id,
	player_id,
	player_opponent_id,
	round_id,	
	round_is_ct	
FROM games
WHERE game_id = ?
`
