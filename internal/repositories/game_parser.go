package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

type GameParserRepository struct {
	flagSourceDir string
	files         []string
	mu            sync.Mutex
}

func NewGameParserRepository(flagSourceDir string) *GameParserRepository {
	return &GameParserRepository{flagSourceDir: flagSourceDir}
}

func (r *GameParserRepository) Next(ctx context.Context) (*types.GameParser, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			r.mu.Lock()

			if len(r.files) == 0 {
				entries, err := os.ReadDir(r.flagSourceDir)
				if err != nil {
					r.mu.Unlock()
					logger.Log.Error(err)
					return nil, err
				}

				var files []string
				for _, entry := range entries {
					name := entry.Name()
					if filepath.Ext(name) == ".json" {
						files = append(files, filepath.Join(r.flagSourceDir, name))
					}
				}
				sort.Strings(files)

				if len(files) == 0 {
					r.mu.Unlock()
					time.Sleep(500 * time.Millisecond)
					continue
				}

				r.files = files
			}

			file := r.files[0]
			r.files = r.files[1:]

			r.mu.Unlock()

			data, err := os.ReadFile(file)
			if err != nil {
				logger.Log.Error(err)
				return nil, err
			}

			var game types.GameParser
			if err := json.Unmarshal(data, &game); err != nil {
				logger.Log.Error(err)
				return nil, err
			}

			return &game, nil
		}
	}
}
