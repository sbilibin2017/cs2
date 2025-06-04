package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/sbilibin2017/cs2/internal/types"
)

type LableEncoderRepository struct {
	flagGameValidDir string
	encoding         map[int]int
}

func NewLableEncoderRepository(flagGameValidDir string) *LableEncoderRepository {
	return &LableEncoderRepository{flagGameValidDir: flagGameValidDir}
}

func (r *LableEncoderRepository) GetEncodingMap() map[int]int {
	return r.encoding
}

func (r *LableEncoderRepository) Fit(ctx context.Context, gameIDs []int) error {
	uniquePlayers := make(map[int32]struct{})

	for _, gameID := range gameIDs {
		fileName := fmt.Sprintf("%d.json", gameID)
		path := filepath.Join(r.flagGameValidDir, fileName)

		f, err := os.Open(path)
		if err != nil {
			continue
		}

		var game types.GameParser
		err = json.NewDecoder(f).Decode(&game)
		f.Close()
		if err != nil {
			continue
		}

		for _, stat := range game.Statistics {
			if stat.Player.ID != 0 {
				uniquePlayers[stat.Player.ID] = struct{}{}
			}
		}
	}

	playerIDs := make([]int, 0, len(uniquePlayers))
	for playerID := range uniquePlayers {
		playerIDs = append(playerIDs, int(playerID))
	}
	sort.Ints(playerIDs)

	encoding := make(map[int]int)
	for label, playerID := range playerIDs {
		encoding[playerID] = label
	}

	r.encoding = encoding

	return nil
}

func (r *LableEncoderRepository) Transform(ctx context.Context, playerID int) (int, error) {
	if r.encoding == nil {
		return -1, fmt.Errorf("encoding not initialized, call Fit first")
	}

	label, ok := r.encoding[playerID]
	if !ok {
		return -1, nil
	}

	return label, nil
}
