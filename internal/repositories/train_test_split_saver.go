package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sbilibin2017/cs2/internal/types"
)

type TrainTestSplitSaverRepository struct {
	flagTrainTestSplitFilePath string
}

func NewTrainTestSplitSaverRepository(flagTrainTestSplitFilePath string) *TrainTestSplitSaverRepository {
	return &TrainTestSplitSaverRepository{flagTrainTestSplitFilePath: flagTrainTestSplitFilePath}
}

func (r *TrainTestSplitSaverRepository) Save(ctx context.Context, split types.TrainTestSplit) error {
	// Ensure the directory exists
	dir := filepath.Dir(r.flagTrainTestSplitFilePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(r.flagTrainTestSplitFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(split); err != nil {
		return err
	}

	return nil
}
