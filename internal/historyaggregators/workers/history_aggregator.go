package workers

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/sbilibin2017/cs2/internal/historyaggregators/types"
	"github.com/sbilibin2017/cs2/internal/logger"
)

type TrainTestSplitLastGetter interface {
	GetLast(ctx context.Context) (*types.TrainTestSplit, error)
}

type GameGetterByID interface {
	Get(ctx context.Context, gameID int32) ([]types.GameDB, error)
}

type HistoryAggregatorPlayer interface {
	Aggregate(ctx context.Context, beginAt time.Time, playerID int32) (*types.Aggregation, error)
}

type HistoryAggregatorPlayerSaver interface {
	Save(ctx context.Context, gameID int32, playerID int32, agg types.Aggregation) error
}

const batchSize = 100

func StartHistoryAggregatorWorker(
	ctx context.Context,
	lg TrainTestSplitLastGetter,
	gg GameGetterByID,
	hap HistoryAggregatorPlayer,
	haps HistoryAggregatorPlayerSaver,
) {
	trainTestSplitCh := generatorTrainTestSplit(ctx, lg)
	gameCh := generatorGame(ctx, gg, trainTestSplitCh)
	featureExtractorParamsCh := generatorFeatureExtractorParams(ctx, gameCh)
	playerAggregationParamsCh := generatorPlayerAggregationParams(ctx, featureExtractorParamsCh)
	aggegatorWorkerPoolCh := playerAggegatorWorkerPool(ctx, hap, playerAggregationParamsCh, 10)
	playerAggregatorResultSaverWorkerPool(ctx, haps, aggegatorWorkerPoolCh, 10)
}

func generatorTrainTestSplit(
	ctx context.Context,
	lg TrainTestSplitLastGetter,
) <-chan int32 {
	out := make(chan int32, batchSize)

	go func() {
		defer close(out)

		// Try to get the last split once, respecting context cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}

		split, err := lg.GetLast(ctx)
		if err != nil {
			logger.Log.Error(err)
			return
		}

		if split != nil {
			for _, id := range split.TrainGameIDs {
				select {
				case <-ctx.Done():
					return
				case out <- id:
				}
			}
		}
	}()

	return out
}

func generatorGame(
	ctx context.Context,
	gg GameGetterByID,
	in <-chan int32,
) <-chan []types.GameDB {
	out := make(chan []types.GameDB, batchSize)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case gameID, ok := <-in:
				if !ok {
					return
				}

				games, err := gg.Get(ctx, gameID)
				if err != nil {
					logger.Log.Error(err)
					continue
				}

				out <- games
			}
		}
	}()

	return out
}

func generatorFeatureExtractorParams(
	ctx context.Context,
	in <-chan []types.GameDB,
) <-chan types.FeatureExtractorParams {
	out := make(chan types.FeatureExtractorParams, batchSize)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case games, ok := <-in:
				if !ok {
					return
				}
				if len(games) == 0 {
					continue
				}
				params := convertGameDBToFeatureExtractorParams(games)
				select {
				case <-ctx.Done():
					return
				case out <- params:
				}
			}
		}
	}()

	return out
}

func convertGameDBToFeatureExtractorParams(games []types.GameDB) types.FeatureExtractorParams {
	var params types.FeatureExtractorParams

	first := games[0]

	params.BeginAt = first.BeginAt
	params.GameID = first.GameID
	params.LeagueID = first.LeagueID
	params.SerieID = first.SerieID
	params.TournamentID = first.TournamentID
	params.TierID = first.TierID
	params.MapID = first.MapID

	teamSet := map[int32]struct{}{}
	for _, g := range games {
		teamSet[g.TeamID] = struct{}{}
		teamSet[g.TeamOpponentID] = struct{}{}
	}

	teams := make([]int32, 0, 2)
	for team := range teamSet {
		teams = append(teams, team)
	}
	sort.Slice(teams, func(i, j int) bool { return teams[i] < teams[j] })
	params.TeamIDs[0] = teams[0]
	params.TeamIDs[1] = teams[1]

	playerSet := map[int32]map[int32]struct{}{teams[0]: {}, teams[1]: {}}
	for _, g := range games {
		playerSet[g.TeamID][g.PlayerID] = struct{}{}
		playerSet[g.TeamOpponentID][g.PlayerOpponentID] = struct{}{}
	}

	playersTeam1 := make([]int32, 0, len(playerSet[teams[0]]))
	for p := range playerSet[teams[0]] {
		playersTeam1 = append(playersTeam1, p)
	}
	sort.Slice(playersTeam1, func(i, j int) bool { return playersTeam1[i] < playersTeam1[j] })

	playersTeam2 := make([]int32, 0, len(playerSet[teams[1]]))
	for p := range playerSet[teams[1]] {
		playersTeam2 = append(playersTeam2, p)
	}
	sort.Slice(playersTeam2, func(i, j int) bool { return playersTeam2[i] < playersTeam2[j] })

	fillFive := func(players []int32) [5]int32 {
		var arr [5]int32
		for i := 0; i < 5; i++ {
			arr[i] = players[i]
		}
		return arr
	}

	team1Players := fillFive(playersTeam1)
	team2Players := fillFive(playersTeam2)

	for i := 0; i < 5; i++ {
		params.PlayerIDs[i] = team1Players[i]
		params.PlayerIDs[i+5] = team2Players[i]
	}

	params.StartCTTeamID = 0
	for _, g := range games {
		if g.RoundIsCT == 1 {
			params.StartCTTeamID = g.TeamID
			break
		}
	}

	return params
}

type playerAggregationParam struct {
	GameID   int32
	BeginAt  time.Time
	PlayerID int32
}

func generatorPlayerAggregationParams(
	ctx context.Context,
	in <-chan types.FeatureExtractorParams,
) <-chan playerAggregationParam {
	out := make(chan playerAggregationParam, batchSize)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case params, ok := <-in:
				if !ok {
					return
				}

				for _, pid := range params.PlayerIDs {
					out <- playerAggregationParam{
						GameID:   params.GameID,
						BeginAt:  params.BeginAt,
						PlayerID: pid,
					}
				}
			}
		}
	}()

	return out
}

type playerAggregatorWorkerResult struct {
	GameID   int32
	PlayerID int32
	types.Aggregation
}

func playerAggegatorWorkerPool(
	ctx context.Context,
	hap HistoryAggregatorPlayer,
	in <-chan playerAggregationParam,
	workerCount int32,
) <-chan playerAggregatorWorkerResult {
	out := make(chan playerAggregatorWorkerResult, batchSize)
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case param, ok := <-in:
				if !ok {
					return
				}

				agg, err := hap.Aggregate(ctx, param.BeginAt, param.PlayerID)
				if err != nil || agg == nil {
					continue
				}

				select {
				case <-ctx.Done():
					return
				case out <- playerAggregatorWorkerResult{
					GameID:      param.GameID,
					PlayerID:    param.PlayerID,
					Aggregation: *agg,
				}:
				}
			}
		}
	}

	wg.Add(int(workerCount))
	for i := int32(0); i < workerCount; i++ {
		go worker()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func playerAggregatorResultSaverWorkerPool(
	ctx context.Context,
	hap HistoryAggregatorPlayerSaver,
	in <-chan playerAggregatorWorkerResult,
	workerCount int32,
) {
	var wg sync.WaitGroup

	for i := int32(0); i < workerCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case result, ok := <-in:
					if !ok {
						return
					}

					err := hap.Save(ctx, result.GameID, result.PlayerID, result.Aggregation)
					if err != nil {
						// Optional: log or handle save error
						continue
					}
				}
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
}
