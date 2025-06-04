package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/sbilibin2017/cs2/internal/types"
)

const testSize = 100

type TrainTestSplitterRepository struct {
	flagGameValidDir string
}

func NewTrainTestSplitterRepository(flagGameValidDir string) *TrainTestSplitterRepository {
	return &TrainTestSplitterRepository{flagGameValidDir: flagGameValidDir}
}

func (r *TrainTestSplitterRepository) Split(ctx context.Context) (*types.TrainTestSplit, error) {
	files, err := os.ReadDir(r.flagGameValidDir)
	if err != nil {
		return nil, err
	}

	var gamesMeta []types.GameMeta

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(r.flagGameValidDir, file.Name())

		f, err := os.Open(path)
		if err != nil {
			continue
		}

		var gameMeta types.GameMeta
		if err := json.NewDecoder(f).Decode(&gameMeta); err == nil &&
			!gameMeta.BeginAt.IsZero() && gameMeta.ID != 0 {
			gamesMeta = append(gamesMeta, gameMeta)
		}
		f.Close()
	}

	sort.Slice(gamesMeta, func(i, j int) bool {
		return gamesMeta[i].BeginAt.Before(gamesMeta[j].BeginAt)
	})

	var trainIDs, testIDs []int
	if len(gamesMeta) > testSize {
		trainGames := gamesMeta[:len(gamesMeta)-testSize]
		testGames := gamesMeta[len(gamesMeta)-testSize:]

		for _, g := range trainGames {
			trainIDs = append(trainIDs, int(g.ID))
		}
		for _, g := range testGames {
			testIDs = append(testIDs, int(g.ID))
		}
	} else {
		for _, g := range gamesMeta {
			testIDs = append(testIDs, int(g.ID))
		}
	}

	return &types.TrainTestSplit{
		TrainIDs: trainIDs,
		TestIDs:  testIDs,
	}, nil
}
