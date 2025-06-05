package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

func StartGameFlattennerWorker(
	ctx context.Context,
	gamesRawDir string,
	gamesFlattenDir string,
) {
	rawGamesCh := generatorGameRaw(ctx, gamesRawDir)
	validGamesCh := validatorGame(ctx, rawGamesCh)
	flattenedGamesCh := flattennerGame(ctx, validGamesCh)
	saverGameFlatten(ctx, gamesFlattenDir, flattenedGamesCh)
}

func generatorGameRaw(
	ctx context.Context,
	gamesRawDir string,
) <-chan types.GameRaw {
	out := make(chan types.GameRaw, 100)

	nextGameRaw := nextGameRawGenerator(ctx, gamesRawDir)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				game, err := nextGameRaw()
				if err != nil {
					logger.Log.Error(err)
					time.Sleep(500 * time.Millisecond)
					continue
				}
				if game == nil {
					time.Sleep(500 * time.Millisecond)
					continue
				}

				out <- *game
			}
		}
	}()

	return out
}

func nextGameRawGenerator(
	ctx context.Context,
	gamesRawDir string,
) func() (*types.GameRaw, error) {
	var (
		mu    sync.Mutex
		seen  = make(map[string]bool)
		files []fs.DirEntry
		index int
	)

	updateFiles := func() {
		newFiles, err := os.ReadDir(gamesRawDir)
		if err != nil {
			logger.Log.Errorf("failed to read dir: %v", err)
			return
		}

		files = nil
		for _, f := range newFiles {
			if !f.IsDir() && !seen[f.Name()] {
				files = append(files, f)
			}
		}
		index = 0
	}

	updateFiles()

	return func() (*types.GameRaw, error) {
		mu.Lock()
		defer mu.Unlock()

		for {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			if index >= len(files) {
				time.Sleep(1 * time.Second)
				updateFiles()
				continue
			}

			entry := files[index]
			index++

			filename := entry.Name()
			path := filepath.Join(gamesRawDir, filename)

			content, err := os.ReadFile(path)
			if err != nil {
				logger.Log.Errorf("failed to read file %s: %v", path, err)
				continue
			}

			game := &types.GameRaw{}
			if err := json.Unmarshal(content, game); err != nil {
				logger.Log.Errorf("invalid game file %s: %v", path, err)
				continue
			}

			seen[filename] = true
			return game, nil
		}
	}
}

func validatorGame(
	ctx context.Context,
	in <-chan types.GameRaw,
) <-chan types.GameRaw {
	out := make(chan types.GameRaw, 100)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case game, ok := <-in:
				if !ok {
					return
				}

				if validateGame(game) {
					out <- game
				}
			}
		}
	}()

	return out
}

func validateGame(game types.GameRaw) bool {
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

func flattennerGame(
	ctx context.Context,
	in <-chan types.GameRaw,
) <-chan []types.GameDB {
	out := make(chan []types.GameDB, 100)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case game, ok := <-in:
				if !ok {
					return
				}
				out <- flattenGame(game)
			}
		}
	}()

	return out
}

func flattenGame(game types.GameRaw) []types.GameDB {
	teamPlayers := make(map[int][]int)
	for _, s := range game.Statistics {
		teamPlayers[int(s.Team.ID)] = append(teamPlayers[int(s.Team.ID)], int(s.Player.ID))
	}

	playerStats := make(map[int]types.StatisticRaw)
	for _, s := range game.Statistics {
		playerStats[int(s.Player.ID)] = s
	}

	var teamIDs []int
	for k := range teamPlayers {
		teamIDs = append(teamIDs, k)
	}

	teamOpponent := make(map[int]int)
	teamOpponent[teamIDs[0]] = teamIDs[1]
	teamOpponent[teamIDs[1]] = teamIDs[0]

	teamPair1 := []int{teamIDs[0], teamIDs[1]}
	teamPair2 := []int{teamIDs[1], teamIDs[0]}
	teamPairs := [][]int{teamPair1, teamPair2}

	tierMap := map[string]int64{"s": 1, "a": 2, "b": 3, "c": 4, "d": 5}
	outcomeMap := map[string]int64{"exploded": 1, "defused": 2, "timeout": 3, "eliminated": 4}

	var gamesDB []types.GameDB
	for _, r := range game.Rounds {
		for _, teamPair := range teamPairs {
			tID := teamPair[0]
			tOppID := teamPair[1]
			tpsID := teamPlayers[tID]
			topppsID := teamPlayers[tOppID]
			for _, pID := range tpsID {
				for _, pOppID := range topppsID {
					tier, ok := tierMap[game.Match.Serie.Tier]
					if !ok {
						tier = int64(-1)
					}
					outcome, ok := outcomeMap[r.Outcome]
					if !ok {
						outcome = int64(-1)
					}
					gameDB := types.GameDB{
						GameID:  int64(game.ID),
						BeginAt: game.BeginAt,

						LeagueID:     int64(game.Match.League.ID),
						SerieID:      int64(game.Match.Serie.ID),
						TournamentID: int64(game.Match.Tournament.ID),
						TierID:       tier,

						MapID: int64(game.Map.ID),

						TeamID:           int64(tID),
						PlayerID:         int64(pID),
						TeamOpponentID:   int64(tOppID),
						PlayerOpponentID: int64(pOppID),

						RoundID:        int64(r.Round),
						RoundOutcomeID: outcome,
						RoundCTID:      int64(r.Ct),
						RoundWinnerID:  int64(r.WinnerTeam),

						Kills:          int64(playerStats[pID].Kills),
						Deaths:         int64(playerStats[pID].Deaths),
						Assists:        int64(playerStats[pID].Assists),
						Headshots:      int64(playerStats[pID].Headshots),
						FlashAssists:   int64(playerStats[pID].FlashAssists),
						FirstKillsDiff: int64(playerStats[pID].FirstKillsDiff),
						KDDiff:         int64(playerStats[pID].KDDiff),
						Adr:            float64(playerStats[pID].Adr),
						Kast:           float64(playerStats[pID].Kast),
						Rating:         float64(playerStats[pID].Rating),
					}

					gamesDB = append(gamesDB, gameDB)
				}
			}
		}
	}

	return gamesDB
}

func saverGameFlatten(
	ctx context.Context,
	pathToSaveDir string,
	in <-chan []types.GameDB,
) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case gamesSlice, ok := <-in:
				if !ok {
					return
				}

				if len(gamesSlice) == 0 {
					continue
				}

				gameID := gamesSlice[0].GameID
				filename := fmt.Sprintf("%d.json", gameID)
				fullPath := filepath.Join(pathToSaveDir, filename)

				data, err := json.MarshalIndent(gamesSlice, "", "  ")
				if err != nil {
					logger.Log.Errorf("failed to marshal games to JSON: %v", err)
					continue
				}

				if err := os.WriteFile(fullPath, data, 0644); err != nil {
					logger.Log.Errorf("failed to write file %s: %v", fullPath, err)
					continue
				}

				logger.Log.Infof("saved %d games for game_id %d to %s", len(gamesSlice), gameID, fullPath)
			}
		}
	}()
}
