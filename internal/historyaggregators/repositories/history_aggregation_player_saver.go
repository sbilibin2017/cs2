package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sbilibin2017/cs2/internal/historyaggregators/types"
	"github.com/sbilibin2017/cs2/internal/logger"
)

type HistoryAggregatorPlayerSaverRepository struct {
	flagAggregationSaveDir string
}

func NewHistoryAggregatorPlayerSaverRepository(flagAggregationSaveDir string) *HistoryAggregatorPlayerSaverRepository {
	return &HistoryAggregatorPlayerSaverRepository{flagAggregationSaveDir: flagAggregationSaveDir}
}

func (r *HistoryAggregatorPlayerSaverRepository) Save(
	ctx context.Context,
	gameID int32,
	playerID int32,
	agg types.Aggregation,
) error {
	dirPath := filepath.Join(r.flagAggregationSaveDir, fmt.Sprintf("%d", gameID))
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		logger.Log.Error(err)
		return err
	}

	filePath := filepath.Join(dirPath, fmt.Sprintf("%d.json", playerID))

	file, err := os.Create(filePath)
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(agg); err != nil {
		logger.Log.Error(err)
		return err
	}

	return nil
}
