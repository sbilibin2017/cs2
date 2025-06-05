package types

type AggregationResult struct {
	KillsTotal        float64 `json:"kills_total"`
	DeathsTotal       float64 `json:"deaths_total"`
	AssistsTotal      float64 `json:"assists_total"`
	HeadshotsTotal    float64 `json:"headshots_total"`
	FlashAssistsTotal float64 `json:"flash_assists_total"`

	KillsPerRound        float64 `json:"kills_per_round"`
	DeathsPerRound       float64 `json:"deaths_per_round"`
	AssistsPerRound      float64 `json:"assists_per_round"`
	HeadshotsPerRound    float64 `json:"headshots_per_round"`
	FlashAssistsPerRound float64 `json:"flash_assists_per_round"`

	KillsPerGame          float64 `json:"kills_per_game"`
	DeathsPerGame         float64 `json:"deaths_per_game"`
	AssistsPerGame        float64 `json:"assists_per_game"`
	HeadshotsPerGame      float64 `json:"headshots_per_game"`
	FlashAssistsPerGame   float64 `json:"flash_assists_per_game"`
	FirstKillsDiffPerGame float64 `json:"first_kills_diff_per_game"`
	KDdiffPerGame         float64 `json:"kd_diff_per_game"`
	ADRPerGame            float64 `json:"adr_per_game"`
	KASTPerGame           float64 `json:"kast_per_game"`
	RatingPerGame         float64 `json:"rating_per_game"`
}
