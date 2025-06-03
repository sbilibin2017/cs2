package types

import (
	"time"

	"github.com/google/uuid"
)

type GameDB struct {
	ID uuid.UUID `json:"id"`

	BeginAt time.Time `json:"begin_at"`

	GameID int32 `json:"game_id"`

	LeagueID     int32 `json:"league_id"`
	SerieID      int32 `json:"serie_id"`
	TournamentID int32 `json:"tournament_id"`
	TierID       int32 `json:"tier_id"`
	MapID        int32 `json:"map_id"`

	TeamID           int32 `json:"team_id" `
	TeamOpponentID   int32 `json:"team_opponent_id"`
	PlayerID         int32 `json:"player_id"`
	PlayerOpponentID int32 `json:"player_opponent_id"`

	Kills          int32   `json:"kills"`
	Deaths         int32   `json:"deaths"`
	Assists        int32   `json:"assists"`
	Headshots      int32   `json:"headshots"`
	FlashAssists   int32   `json:"flash_assists"`
	FirstKillsDiff int32   `json:"first_kills_diff"`
	KDDiff         int32   `json:"k_d_diff"`
	Adr            float32 `json:"adr"`
	Kast           float32 `json:"kast"`
	Rating         float32 `json:"rating"`

	RoundID        int32 `json:"round_id"`
	RoundOutcomeID int32 `json:"round_outcome_id"`
	RoundIsCT      int32 `json:"round_is_ct"`
	RoundWin       int32 `json:"round_win"`

	UpdatedAt time.Time `json:"updated_at"`
}
