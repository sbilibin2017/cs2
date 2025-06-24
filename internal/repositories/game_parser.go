package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/sbilibin2017/cs2/internal/types"
)

type GameParserOption func(*GameParserRepository)

type GameParserRepository struct {
	pathToDir string
	mu        sync.RWMutex
	files     []string
	index     int
}

func WithPathToDir(path string) GameParserOption {
	return func(r *GameParserRepository) {
		r.pathToDir = path
	}
}

func NewGameParserRepository(opts ...GameParserOption) *GameParserRepository {
	repo := &GameParserRepository{
		files: make([]string, 0),
		index: 0,
	}

	for _, opt := range opts {
		opt(repo)
	}

	return repo
}

func (repo *GameParserRepository) Next(ctx context.Context) (*types.GameParser, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if len(repo.files) == 0 || repo.index == 0 {
		entries, err := os.ReadDir(repo.pathToDir)
		if err != nil {
			return nil, err
		}

		repo.files = repo.files[:0]
		for _, entry := range entries {
			if !entry.IsDir() {
				repo.files = append(repo.files, filepath.Join(repo.pathToDir, entry.Name()))
			}
		}

		if len(repo.files) == 0 {
			return nil, errors.New("no games found in directory")
		}
	}

	filePath := repo.files[repo.index]
	repo.index = (repo.index + 1) % len(repo.files)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var game types.GameParser
	if err := json.Unmarshal(data, &game); err != nil {
		return nil, err
	}

	return &game, nil
}
