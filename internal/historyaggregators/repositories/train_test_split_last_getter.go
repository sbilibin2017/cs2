package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/sbilibin2017/cs2/internal/historyaggregators/types"
	"github.com/sbilibin2017/cs2/internal/logger"
)

type TrainTestSplitGetterByHashRepository struct {
	flagTrainTestSplitDir string
}

func NewTrainTestSplitGetterByHashRepository(flagTrainTestSplitDir string) *TrainTestSplitGetterByHashRepository {
	return &TrainTestSplitGetterByHashRepository{flagTrainTestSplitDir: flagTrainTestSplitDir}
}

func (r *TrainTestSplitGetterByHashRepository) GetLast(
	ctx context.Context,
) (*types.TrainTestSplit, error) {
	var (
		latestSplit     *types.TrainTestSplit
		latestUpdatedAt time.Time
	)

	err := filepath.Walk(r.flagTrainTestSplitDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			logger.Log.Error("failed to open file:", err)
			return nil
		}
		defer file.Close()

		var split types.TrainTestSplit
		if err := json.NewDecoder(file).Decode(&split); err != nil {
			logger.Log.Error("failed to decode file:", err)
			return nil
		}

		if split.UpdatedAt.After(latestUpdatedAt) {
			latestSplit = &split
			latestUpdatedAt = split.UpdatedAt
		}

		return nil
	})

	if err != nil {
		logger.Log.Error("error walking files:", err)
		return nil, err
	}

	if latestSplit == nil {
		return nil, os.ErrNotExist
	}

	return latestSplit, nil
}
