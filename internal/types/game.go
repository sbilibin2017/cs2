package types

import (
	"time"
)

type MapRaw struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type LeagueRaw struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type SerieRaw struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Tier string `json:"tier"`
}

type TournamentRaw struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	PrizePool string `json:"prizepool"`
}

type MatchRaw struct {
	League     LeagueRaw     `json:"league"`
	Serie      SerieRaw      `json:"serie"`
	Tournament TournamentRaw `json:"tournament"`
}

type TeamRaw struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type PlayerRaw struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Hometown string `json:"hometown,omitempty"`
	Birthday string `json:"birthday,omitempty"`
}

type StatisticRaw struct {
	Team           TeamRaw   `json:"team"`
	Player         PlayerRaw `json:"player"`
	Kills          int32     `json:"kills"`
	Deaths         int32     `json:"deaths"`
	Assists        int32     `json:"assists"`
	Headshots      int32     `json:"headshots"`
	FlashAssists   int32     `json:"flash_assists"`
	FirstKillsDiff int32     `json:"first_kills_diff"`
	KDDiff         int32     `json:"k_d_diff"`
	Adr            float32   `json:"adr"`
	Kast           float32   `json:"kast"`
	Rating         float32   `json:"rating"`
}

type RoundRaw struct {
	Round      int32  `json:"round"`
	Ct         int32  `json:"ct"`
	Terrorists int32  `json:"terrorists"`
	WinnerTeam int32  `json:"winner_team"`
	Outcome    string `json:"outcome"`
}

type GameRaw struct {
	ID         int32          `json:"id"`
	BeginAt    time.Time      `json:"begin_at"`
	Match      MatchRaw       `json:"match"`
	Map        MapRaw         `json:"map"`
	Statistics []StatisticRaw `json:"players"`
	Rounds     []RoundRaw     `json:"rounds"`
}

type GameDB struct {
	GameID  int64     `json:"game_id"`
	BeginAt time.Time `json:"begin_at"`

	LeagueID     int64 `json:"league_id"`
	SerieID      int64 `json:"serie_id"`
	TournamentID int64 `json:"tournament_id"`
	TierID       int64 `json:"tier_id"`

	MapID int64 `json:"map_id"`

	TeamID           int64 `json:"team_id"`
	PlayerID         int64 `json:"player_id"`
	TeamOpponentID   int64 `json:"team_opponent_id"`
	PlayerOpponentID int64 `json:"player_opponent_id"`

	RoundID        int64 `json:"round_id"`
	RoundOutcomeID int64 `json:"round_outcome_id"`
	RoundCTID      int64 `json:"round_ct_id"`
	RoundWinnerID  int64 `json:"round_winner_id"`

	Kills          int64   `json:"kills"`
	Deaths         int64   `json:"deaths"`
	Assists        int64   `json:"assists"`
	Headshots      int64   `json:"headshots"`
	FlashAssists   int64   `json:"flash_assists"`
	FirstKillsDiff int64   `json:"first_kills_diff"`
	KDDiff         int64   `json:"k_d_diff"`
	Adr            float64 `json:"adr"`
	Kast           float64 `json:"kast"`
	Rating         float64 `json:"rating"`
}
