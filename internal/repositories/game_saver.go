package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sbilibin2017/cs2/internal/types"
)

type GameSaverRepository struct {
	flagGameValidDir string
}

func NewGameSaverRepository(flagGameValidDir string) *GameSaverRepository {
	return &GameSaverRepository{flagGameValidDir: flagGameValidDir}
}

func (r *GameSaverRepository) Save(ctx context.Context, game types.GameParser) error {
	if err := os.MkdirAll(r.flagGameValidDir, os.ModePerm); err != nil {
		return err
	}

	fileName := fmt.Sprintf("%d.json", game.ID)
	filePath := filepath.Join(r.flagGameValidDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(game); err != nil {
		return err
	}

	return nil
}
