package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/traintestsplitter/types"
)

type TrainTestSplitSaverRepository struct {
	flagDestinationDir string
}

func NewTrainTestSplitSaverRepository(flagDestinationDir string) *TrainTestSplitSaverRepository {
	return &TrainTestSplitSaverRepository{flagDestinationDir: flagDestinationDir}
}

func (r *TrainTestSplitSaverRepository) Save(ctx context.Context, split types.TrainTestSplit) error {
	filename := filepath.Join(r.flagDestinationDir, fmt.Sprintf("%s.json", split.Hash))
	file, err := os.Create(filename)
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(split); err != nil {
		logger.Log.Error(err)
		return err
	}

	return nil
}
