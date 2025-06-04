package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sbilibin2017/cs2/internal/types"
)

type GameLoaderRepository struct {
	flagSourceDir string
}

func NewGameLoaderRepository(flagSourceDir string) *GameLoaderRepository {
	return &GameLoaderRepository{flagSourceDir: flagSourceDir}
}

func (r *GameLoaderRepository) Load(ctx context.Context, gameID int) (*types.GameParser, error) {
	fileName := fmt.Sprintf("%d.json", gameID)
	filePath := filepath.Join(r.flagSourceDir, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open game file %s: %w", filePath, err)
	}
	defer file.Close()

	var game types.GameParser
	if err := json.NewDecoder(file).Decode(&game); err != nil {
		return nil, fmt.Errorf("failed to decode game file %s: %w", filePath, err)
	}

	return &game, nil
}
