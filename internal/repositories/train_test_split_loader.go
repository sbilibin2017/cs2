package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sbilibin2017/cs2/internal/types"
)

type TrainTestSplitLoaderRepository struct {
	flagTrainTestSplitFilePath string
}

func NewTrainTestSplitLoaderRepository(flagTrainTestSplitFilePath string) *TrainTestSplitLoaderRepository {
	return &TrainTestSplitLoaderRepository{flagTrainTestSplitFilePath: flagTrainTestSplitFilePath}
}

func (r *TrainTestSplitLoaderRepository) Load(ctx context.Context) (*types.TrainTestSplit, error) {
	file, err := os.Open(r.flagTrainTestSplitFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open train-test split file: %w", err)
	}
	defer file.Close()

	var split types.TrainTestSplit
	if err := json.NewDecoder(file).Decode(&split); err != nil {
		return nil, fmt.Errorf("failed to decode train-test split JSON: %w", err)
	}

	return &split, nil
}
