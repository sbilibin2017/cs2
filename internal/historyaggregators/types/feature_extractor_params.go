package types

import "time"

type FeatureExtractorParams struct {
	BeginAt time.Time `json:"begin_at"`

	GameID int32 `json:"game_id"`

	LeagueID     int32 `json:"league_id"`
	SerieID      int32 `json:"serie_id"`
	TournamentID int32 `json:"tournament_id"`
	TierID       int32 `json:"tier_id"`
	MapID        int32 `json:"map_id"`

	TeamIDs   [2]int32  `json:"team_ids"`
	PlayerIDs [10]int32 `json:"player_ids"`

	StartCTTeamID int32 `json:"start_ct_team_id"`
}
