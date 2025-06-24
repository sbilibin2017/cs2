package types

import "time"

type LeagueParser struct {
	ID int64 `json:"id"`
}

type SerieParser struct {
	ID       int64  `json:"id"`
	Tier     string `json:"tier"`
	LeagueID int64  `json:"league_id"`
}

type TournamentParser struct {
	ID      int64 `json:"id"`
	SerieID int64 `json:"serie_id"`
}

type MatchParser struct {
	League     LeagueParser     `json:"league"`
	Serie      SerieParser      `json:"serie"`
	Tournament TournamentParser `json:"tournament"`
}

type MapParser struct {
	ID int64 `json:"id"`
}

type TeamParser struct {
	ID int64 `json:"id"`
}

type PlayerParser struct {
	ID int64 `json:"id"`
}

type PlayerStatisticParser struct {
	Team           TeamParser   `json:"team"`
	Player         PlayerParser `json:"player"`
	Kills          float64      `json:"kills"`
	Deaths         float64      `json:"deaths"`
	Assists        float64      `json:"assists"`
	Headshots      float64      `json:"headshots"`
	FlashAssists   float64      `json:"flash_assists"`
	KDDiff         float64      `json:"kd_diff"`
	FirstKillsDiff float64      `json:"first_kills_diff"`
	ADR            float64      `json:"adr"`
	Kast           float64      `json:"kast"`
	Rating         float64      `json:"rating"`
}

type RoundParser struct {
	Round      int64  `json:"round"`
	CT         int64  `json:"ct"`
	T          int64  `json:"terrorists"`
	WinnerTeam int64  `json:"winner_team"`
	Outcome    string `json:"outcome"`
}

type GameParser struct {
	ID      int64                   `json:"id"`
	BeginAt time.Time               `json:"begin_at"`
	Match   MatchParser             `json:"match"`
	Map     MapParser               `json:"map"`
	Players []PlayerStatisticParser `json:"players"`
	Rounds  []RoundParser           `json:"rounds"`
}

type GameDB struct {
	GameID  int64     `json:"game_id"`
	BeginAt time.Time `json:"begin_at"`

	LeagueID     int64 `json:"league_id"`
	SerieID      int64 `json:"serie_id"`
	TierID       int64 `json:"tier_id"`
	TournamentID int64 `json:"tournament_id"`

	MapID int64 `json:"map_id"`

	TeamID           int64 `json:"team_id"`
	TeamOpponentID   int64 `json:"team_opponent_id"`
	PlayerID         int64 `json:"player_id"`
	PlayerOpponentID int64 `json:"player_opponent_id"`

	Kills          int64   `json:"kills"`
	Deaths         int64   `json:"deaths"`
	Assists        int64   `json:"assists"`
	Headshots      int64   `json:"headshots"`
	FlashAssists   int64   `json:"flash_assists"`
	KDDiff         float64 `json:"k_d_diff"`
	FirstKillsDiff float64 `json:"first_kills_diff"`
	ADR            float64 `json:"adr"`
	Kast           float64 `json:"kast"`
	Rating         float64 `json:"rating"`

	RoundID        int64 `json:"round_id"`
	RoundOutcomeID int64 `json:"round_outcome_id"`
	RoundWin       int64 `json:"round_win"`
}
