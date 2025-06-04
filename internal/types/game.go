package types

import (
	"time"
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

type GameMeta struct {
	ID      int32     `json:"id"`
	BeginAt time.Time `json:"begin_at"`
}
