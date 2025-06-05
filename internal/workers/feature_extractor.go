package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

type FeatureExtractor interface {
	Extract(
		ctx context.Context,
		params types.FeaturePlayerParams,
	) (map[string]types.FeaturePlayer, error)
}

func StartFeatureExtractorWorker(
	ctx context.Context,
	featuresDir string,
	conn clickhouse.Conn,
	fe FeatureExtractor,
) {
	featureExtractorGameIDCh := generatorFeatureExtractorGameID(ctx, conn)
	featurePlayerParamsCh := featurePlayerParamsExtractor(ctx, conn, featureExtractorGameIDCh)
	playerFeaturesCh := playerFeatures(ctx, featurePlayerParamsCh, fe, 20)
	savePlayerFeatures(ctx, featuresDir, playerFeaturesCh)
}

func generatorFeatureExtractorGameID(
	ctx context.Context,
	conn clickhouse.Conn,
) <-chan int64 {
	ch := make(chan int64, 100)

	go func() {
		defer close(ch)

		query := `
			SELECT DISTINCT game_id 
			FROM games
			ORDER BY begin_at DESC
		`
		const pollInterval = 10 * time.Second

		for {
			rows, err := conn.Query(ctx, query)
			if err != nil {
				log.Printf("failed to query game IDs: %v", err)
				select {
				case <-time.After(pollInterval):
					continue
				case <-ctx.Done():
					return
				}
			}

			for rows.Next() {
				var gameID int64
				if err := rows.Scan(&gameID); err != nil {
					log.Printf("failed to scan game_id: %v", err)
					continue
				}

				select {
				case ch <- gameID:
				case <-ctx.Done():
					rows.Close()
					return
				}
			}

			rows.Close()

			select {
			case <-time.After(pollInterval):
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

type featurePlayerParams struct {
	GameID int64
	types.FeaturePlayerParams
}

func featurePlayerParamsExtractor(
	ctx context.Context,
	conn clickhouse.Conn,
	gameIDCh <-chan int64,
) <-chan featurePlayerParams {
	outCh := make(chan featurePlayerParams, 100)

	go func() {
		defer close(outCh)

		for {
			select {
			case <-ctx.Done():
				return
			case gameID, ok := <-gameIDCh:
				if !ok {
					return
				}

				beginAt, err := fetchGameBeginAt(ctx, conn, gameID)
				if err != nil {
					continue
				}

				teamIDs, err := fetchTeamIDs(ctx, conn, gameID)
				if err != nil || len(teamIDs) < 2 {
					continue
				}

				players1, err := fetchPlayers(ctx, conn, gameID, teamIDs[0])
				if err != nil {
					continue
				}

				players2, err := fetchPlayers(ctx, conn, gameID, teamIDs[1])
				if err != nil {
					continue
				}

				playerIDs := buildPlayerIDsArray(players1, players2)

				select {
				case outCh <- featurePlayerParams{
					GameID: gameID,
					FeaturePlayerParams: types.FeaturePlayerParams{
						BeginAt:   beginAt,
						PlayerIDs: playerIDs,
					},
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outCh
}

func fetchGameBeginAt(ctx context.Context, conn clickhouse.Conn, gameID int64) (time.Time, error) {
	var beginAt time.Time
	err := conn.QueryRow(ctx, `
		SELECT begin_at
		FROM games
		WHERE game_id = ?
		LIMIT 1
	`, gameID).Scan(&beginAt)
	return beginAt, err
}

func fetchTeamIDs(ctx context.Context, conn clickhouse.Conn, gameID int64) ([]int64, error) {
	rows, err := conn.Query(ctx, `
		SELECT DISTINCT team_id
		FROM games
		WHERE game_id = ?
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teamIDs []int64
	for rows.Next() {
		var teamID int64
		if err := rows.Scan(&teamID); err != nil {
			return nil, err
		}
		teamIDs = append(teamIDs, teamID)
	}
	sort.Slice(teamIDs, func(i, j int) bool { return teamIDs[i] < teamIDs[j] })
	return teamIDs, nil
}

func fetchPlayers(ctx context.Context, conn clickhouse.Conn, gameID, teamID int64) ([]int64, error) {
	rows, err := conn.Query(ctx, `
		SELECT DISTINCT player_id
		FROM games
		WHERE game_id = ? AND team_id = ?
		ORDER BY player_id ASC
	`, gameID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []int64
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			return nil, err
		}
		players = append(players, pid)
	}
	return players, nil
}

func buildPlayerIDsArray(players1, players2 []int64) [10]int64 {
	var playerIDs [10]int64
	for i := 0; i < 5 && i < len(players1); i++ {
		playerIDs[i] = players1[i]
	}
	for i := 0; i < 5 && i < len(players2); i++ {
		playerIDs[5+i] = players2[i]
	}
	return playerIDs
}

func playerFeatures(
	ctx context.Context,
	paramsCh <-chan featurePlayerParams,
	extractor FeatureExtractor,
	workerCount int,
) <-chan featuresPlayer {

	outCh := make(chan featuresPlayer, 100)

	var wg sync.WaitGroup
	wg.Add(workerCount)

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case params, ok := <-paramsCh:
				if !ok {
					return
				}

				features, err := extractor.Extract(ctx, params.FeaturePlayerParams)
				if err != nil {
					logger.Log.Error(err)
					continue
				}

				select {
				case outCh <- featuresPlayer{
					GameID:   params.GameID,
					Features: features,
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	for i := 0; i < workerCount; i++ {
		go worker()
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}

type featuresPlayer struct {
	GameID   int64
	Features map[string]types.FeaturePlayer
}

func savePlayerFeatures(
	ctx context.Context,
	featuresDir string,
	playerFeaturesCh <-chan featuresPlayer,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case features, ok := <-playerFeaturesCh:
			if !ok {
				continue
			}

			err := saveFeatureMap(featuresDir, features)

			if err != nil {
				logger.Log.Error(err)
			}
		}
	}
}

func saveFeatureMap(
	featuresDir string,
	features featuresPlayer,
) error {
	data, err := json.MarshalIndent(features.Features, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	filename := fmt.Sprintf("%d.json", features.GameID)
	filepath := filepath.Join(featuresDir, filename)

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("file create failed: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("file write failed: %w", err)
	}

	return nil
}
