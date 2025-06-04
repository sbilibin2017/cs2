package services

import (
	"context"
	"sort"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

type TrainTestSplitLoader interface {
	Load(ctx context.Context) (*types.TrainTestSplit, error)
}

type LableEncoder interface {
	Fit(ctx context.Context, gameIDs []int) error
	Transform(ctx context.Context, playerID int) (int, error)
}

type GameLoader interface {
	Load(ctx context.Context, gameID int) (*types.GameParser, error)
}

type DatasetMakerService struct {
	le LableEncoder
	sl TrainTestSplitLoader
	gl GameLoader
}

func NewDatasetMakerService(
	le LableEncoder,
	sl TrainTestSplitLoader,
	gl GameLoader,
) *DatasetMakerService {
	return &DatasetMakerService{
		le: le,
		sl: sl,
		gl: gl,
	}
}

func (svc *DatasetMakerService) MakeDataset(
	ctx context.Context,
) ([]types.DatasetRow, error) {
	logger.Log.Info("Loading train/test split...")
	split, err := svc.sl.Load(ctx)
	if err != nil {
		logger.Log.Errorf("Failed to load train/test split: %v", err)
		return nil, err
	}

	logger.Log.Infof("Fitting label encoder with %d games...", len(split.TrainIDs))
	if err := svc.le.Fit(ctx, split.TrainIDs); err != nil {
		logger.Log.Errorf("Failed to fit label encoder: %v", err)
		return nil, err
	}

	var rows []types.DatasetRow

	for _, gameID := range split.TrainIDs {
		logger.Log.Infof("Processing game ID: %d", gameID)

		game, err := svc.gl.Load(ctx, gameID)
		if err != nil {
			logger.Log.Errorf("Failed to load game ID %d: %v", gameID, err)
			return nil, err
		}

		teamPlayers := make(map[int][]int)
		for _, stat := range game.Statistics {
			playerID := int(stat.Player.ID)
			label, err := svc.le.Transform(ctx, playerID)
			if err != nil {
				logger.Log.Errorf("Failed to transform player ID %d: %v", playerID, err)
				return nil, err
			}
			teamID := int(stat.Team.ID)
			teamPlayers[teamID] = append(teamPlayers[teamID], label)
		}

		for teamID := range teamPlayers {
			sort.Ints(teamPlayers[teamID])
		}

		var teamIDs []int
		for id := range teamPlayers {
			teamIDs = append(teamIDs, id)
		}
		sort.Ints(teamIDs)

		roundWins := make(map[int]int)
		for _, round := range game.Rounds {
			roundWins[int(round.WinnerTeam)]++
		}

		team1ID := teamIDs[0]
		team2ID := teamIDs[1]
		target := 0
		if roundWins[team1ID] > roundWins[team2ID] {
			target = 1
		}

		players := append(teamPlayers[team1ID], teamPlayers[team2ID]...)
		row := types.DatasetRow{
			Player1ID:  players[0],
			Player2ID:  players[1],
			Player3ID:  players[2],
			Player4ID:  players[3],
			Player5ID:  players[4],
			Player6ID:  players[5],
			Player7ID:  players[6],
			Player8ID:  players[7],
			Player9ID:  players[8],
			Player10ID: players[9],
			Target:     target,
		}

		rows = append(rows, row)
	}

	logger.Log.Infof("Dataset creation complete: %d rows", len(rows))
	return rows, nil
}
