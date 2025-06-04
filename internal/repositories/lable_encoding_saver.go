package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

type LableEncodingSaverRepository struct {
	flagLableEncodingFilePath string
}

func NewLableEncodingSaverRepository(flagLableEncodingFilePath string) *LableEncodingSaverRepository {
	return &LableEncodingSaverRepository{flagLableEncodingFilePath: flagLableEncodingFilePath}
}

func (r *LableEncodingSaverRepository) Save(ctx context.Context, encoding map[int]int) error {

	dir := filepath.Dir(r.flagLableEncodingFilePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(r.flagLableEncodingFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(encoding); err != nil {
		return err
	}

	return nil
}
