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
	ID uuid.UUID `json:"id" ch:"id"`

	BeginAt time.Time `json:"begin_at" ch:"begin_at"`

	GameID int32 `json:"game_id" ch:"game_id"`

	RoundID        int32 `json:"round_id" ch:"round_id"`
	RoundOutcomeID int32 `json:"outcome" ch:"outcome"`
	RoundIsCT      int32 `json:"is_ct" ch:"is_ct"`

	LeagueID     int32 `json:"league_id" ch:"league_id"`
	SerieID      int32 `json:"serie_id" ch:"serie_id"`
	TournamentID int32 `json:"tournament_id" ch:"tournament_id"`
	TierID       int32 `json:"tier_id" ch:"tier_id"`
	MapID        int32 `json:"map_id" ch:"map_id"`

	TeamID           int32 `json:"team_id" ch:"team_id"`
	TeamOpponentID   int32 `json:"team_opponent_id" ch:"team_opponent_id"`
	PlayerID         int32 `json:"player_id" ch:"player_id"`
	PlayerOpponentID int32 `json:"player_opponent_id" ch:"player_opponent_id"`

	Kills          int32   `json:"kills" ch:"kills"`
	Deaths         int32   `json:"deaths" ch:"deaths"`
	Assists        int32   `json:"assists" ch:"assists"`
	Headshots      int32   `json:"headshots" ch:"headshots"`
	FlashAssists   int32   `json:"flash_assists" ch:"flash_assists"`
	FirstKillsDiff int32   `json:"first_kills_diff" ch:"first_kills_diff"`
	KDDiff         int32   `json:"k_d_diff" ch:"k_d_diff"`
	Adr            float32 `json:"adr" ch:"adr"`
	Kast           float32 `json:"kast" ch:"kast"`
	Rating         float32 `json:"rating" ch:"rating"`

	Win int32 `json:"win" db:"win"`

	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
