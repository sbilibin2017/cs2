package workers

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

func StartGameFlattenLoaderWorker(
	ctx context.Context,
	gamesFlattenDir string,
	conn clickhouse.Conn,
) {
	gamesFlattenCh := generatorGameFlatten(ctx, gamesFlattenDir)
	saverGameDB(ctx, conn, gamesFlattenCh)
}

func generatorGameFlatten(
	ctx context.Context,
	gamesFlattenDir string,
) <-chan []types.GameDB {
	out := make(chan []types.GameDB, 100)

	nextGameFlatten := nextGameFlattenGenerator(ctx, gamesFlattenDir)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				games, err := nextGameFlatten()
				if err != nil {
					logger.Log.Error(err)
					time.Sleep(500 * time.Millisecond)
					continue
				}
				if games == nil {
					time.Sleep(500 * time.Millisecond)
					continue
				}

				out <- games
			}
		}
	}()

	return out
}

func nextGameFlattenGenerator(
	ctx context.Context,
	gamesFlattenDir string,
) func() ([]types.GameDB, error) {
	var (
		mu    sync.Mutex
		seen  = make(map[string]bool)
		files []fs.DirEntry
		index int
	)

	updateFiles := func() {
		newFiles, err := os.ReadDir(gamesFlattenDir)
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

	return func() ([]types.GameDB, error) {
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
			path := filepath.Join(gamesFlattenDir, filename)

			content, err := os.ReadFile(path)
			if err != nil {
				logger.Log.Errorf("failed to read file %s: %v", path, err)
				continue
			}

			var games []types.GameDB
			if err := json.Unmarshal(content, &games); err != nil {
				logger.Log.Errorf("invalid game file %s: %v", path, err)
				continue
			}

			seen[filename] = true
			return games, nil
		}
	}
}

func saverGameDB(
	ctx context.Context,
	conn clickhouse.Conn,
	in <-chan []types.GameDB,
) {
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

			err := func() error {
				batch, err := conn.PrepareBatch(ctx, `
					INSERT INTO games (
						game_id,
						begin_at,
						league_id,
						serie_id,
						tournament_id,
						tier_id,
						map_id,
						team_id,
						player_id,
						team_opponent_id,
						player_opponent_id,
						round_id,
						round_outcome_id,
						round_ct_id,
						round_winner_id,
						kills,
						deaths,
						assists,
						headshots,
						flash_assists,
						first_kills_diff,
						k_d_diff,
						adr,
						kast,
						rating,
						updated_at
					)
				`)
				if err != nil {
					return err
				}

				for _, game := range gamesSlice {
					if err := batch.Append(
						game.GameID,
						game.BeginAt,
						game.LeagueID,
						game.SerieID,
						game.TournamentID,
						game.TierID,
						game.MapID,
						game.TeamID,
						game.PlayerID,
						game.TeamOpponentID,
						game.PlayerOpponentID,
						game.RoundID,
						game.RoundOutcomeID,
						game.RoundCTID,
						game.RoundWinnerID,
						game.Kills,
						game.Deaths,
						game.Assists,
						game.Headshots,
						game.FlashAssists,
						game.FirstKillsDiff,
						game.KDDiff,
						game.Adr,
						game.Kast,
						game.Rating,
						time.Now(), // updated_at
					); err != nil {
						return err
					}
				}

				return batch.Send()
			}()

			if err != nil {
				logger.Log.Errorf("failed to insert games: %v", err)
				continue
			}

			logger.Log.Infof("inserted %d game records", len(gamesSlice))
		}
	}
}
