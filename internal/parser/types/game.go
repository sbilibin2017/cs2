package types

import (
	"time"

	"github.com/google/uuid"
)

type MapParser struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type LeagueParser struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type SerieParser struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Tier string `json:"tier"`
}

type TournamentParser struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	PrizePool string `json:"prizepool"`
}

type MatchParser struct {
	League     LeagueParser     `json:"league"`
	Serie      SerieParser      `json:"serie"`
	Tournament TournamentParser `json:"tournament"`
}

type TeamParser struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type PlayerParser struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Hometown string `json:"hometown,omitempty"`
	Birthday string `json:"birthday,omitempty"`
}

type StatisticParser struct {
	Team           TeamParser   `json:"team"`
	Player         PlayerParser `json:"player"`
	Kills          int32        `json:"kills"`
	Deaths         int32        `json:"deaths"`
	Assists        int32        `json:"assists"`
	Headshots      int32        `json:"headshots"`
	FlashAssists   int32        `json:"flash_assists"`
	FirstKillsDiff int32        `json:"first_kills_diff"`
	KDDiff         int32        `json:"k_d_diff"`
	Adr            float32      `json:"adr"`
	Kast           float32      `json:"kast"`
	Rating         float32      `json:"rating"`
}

type RoundParser struct {
	Round      int32  `json:"round"`
	Ct         int32  `json:"ct"`
	Terrorists int32  `json:"terrorists"`
	WinnerTeam int32  `json:"winner_team"`
	Outcome    string `json:"outcome"`
}

type GameParser struct {
	ID         int32             `json:"id"`
	BeginAt    time.Time         `json:"begin_at"`
	Match      MatchParser       `json:"match"`
	Map        MapParser         `json:"map"`
	Statistics []StatisticParser `json:"players"`
	Rounds     []RoundParser     `json:"rounds"`
}

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
