package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sbilibin2017/cs2/internal/types"
)

type DatasetSaverRepository struct {
	flagPathToDatasetFile string
}

func NewDatasetSaverRepository(flagPathToDatasetFile string) *DatasetSaverRepository {
	return &DatasetSaverRepository{flagPathToDatasetFile: flagPathToDatasetFile}
}

func (r *DatasetSaverRepository) Save(ctx context.Context, dataset []types.DatasetRow) error {
	dir := filepath.Dir(r.flagPathToDatasetFile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(r.flagPathToDatasetFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(dataset); err != nil {
		return err
	}

	return nil
}
