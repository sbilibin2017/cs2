package types

type Aggregation struct {
	KillsTotal    float64 `json:"kills_total"`
	KillsPerGame  float64 `json:"kills_per_game"`
	KillsPerRound float64 `json:"kills_per_round"`

	DeathsTotal    float64 `json:"deaths_total"`
	DeathsPerGame  float64 `json:"deaths_per_game"`
	DeathsPerRound float64 `json:"deaths_per_round"`

	AssistsTotal    float64 `json:"assists_total"`
	AssistsPerGame  float64 `json:"assists_per_game"`
	AssistsPerRound float64 `json:"assists_per_round"`

	HeadshotsTotal    float64 `json:"headshots_total"`
	HeadshotsPerGame  float64 `json:"headshots_per_game"`
	HeadshotsPerRound float64 `json:"headshots_per_round"`

	FlashAssistsTotal    float64 `json:"flash_assists_total"`
	FlashAssistsPerGame  float64 `json:"flash_assists_per_game"`
	FlashAssistsPerRound float64 `json:"flash_assists_per_round"`

	FirstKillsDiffTotal    float64 `json:"first_kills_diff_total"`
	FirstKillsDiffPerGame  float64 `json:"first_kills_diff_per_game"`
	FirstKillsDiffPerRound float64 `json:"first_kills_diff_per_round"`

	KDDiffTotal    float64 `json:"k_d_diff_total"`
	KDDiffPerGame  float64 `json:"k_d_diff_per_game"`
	KDDiffPerRound float64 `json:"k_d_diff_per_round"`

	AdrTotal    float64 `json:"adr_total"`
	AdrPerGame  float64 `json:"adr_per_game"`
	AdrPerRound float64 `json:"adr_per_round"`

	KastTotal    float64 `json:"kast_total"`
	KastPerGame  float64 `json:"kast_per_game"`
	KastPerRound float64 `json:"kast_per_round"`

	RatingTotal    float64 `json:"rating_total"`
	RatingPerGame  float64 `json:"rating_per_game"`
	RatingPerRound float64 `json:"rating_per_round"`

	WinRoundTotal   float64 `json:"win_round_total"`
	WinRoundPerGame float64 `json:"win_round_per_game"`

	WinRoundH1Total   float64 `json:"win_round_h1_total"`
	WinRoundH1PerGame float64 `json:"win_round_h1_per_game"`

	WinRoundH2Total   float64 `json:"win_round_h2_total"`
	WinRoundH2PerGame float64 `json:"win_round_h2_per_game"`

	IsCTRoundTotal   float64 `json:"is_ct_round_total"`
	IsCTRoundPerGame float64 `json:"is_ct_round_per_game"`

	GamesCount  float64 `json:"games_count"`
	RoundsCount float64 `json:"rounds_count"`
}
