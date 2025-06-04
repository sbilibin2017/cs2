package validators

import "github.com/sbilibin2017/cs2/internal/types"

func ValidateGame(game types.GameParser) bool {
	if game.ID == 0 || game.BeginAt.IsZero() {
		return false
	}
	if game.Map.ID == 0 {
		return false
	}
	if game.Match.League.ID == 0 || game.Match.Serie.ID == 0 || game.Match.Tournament.ID == 0 {
		return false
	}

	teams := map[int32]struct{}{}
	for _, stat := range game.Statistics {
		if stat.Team.ID == 0 {
			return false
		}
		teams[stat.Team.ID] = struct{}{}
	}
	if len(teams) != 2 {
		return false
	}

	players := map[int32]struct{}{}
	for _, stat := range game.Statistics {
		if stat.Player.ID == 0 {
			return false
		}
		players[stat.Player.ID] = struct{}{}
	}
	if len(players) != 10 {
		return false
	}

	if len(game.Rounds) == 0 || game.Rounds[0].Round != 1 {
		return false
	}

	for _, r := range game.Rounds {
		if r.Round == 0 || r.Ct == 0 || r.Terrorists == 0 || r.WinnerTeam == 0 {
			return false
		}
	}

	lastRound := game.Rounds[len(game.Rounds)-1]
	return lastRound.Round >= 16
}
