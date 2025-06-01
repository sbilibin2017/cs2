package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

func writeTestGameFile(t *testing.T, dir, filename string, game *types.GameParser) {
	t.Helper()
	data, err := json.Marshal(game)
	require.NoError(t, err)

	filePath := filepath.Join(dir, filename)
	err = os.WriteFile(filePath, data, 0644)
	require.NoError(t, err)
}

func TestGameParserRepository_Next(t *testing.T) {
	logger.Initialize("debug")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tempDir := t.TempDir()

	// Write both game files BEFORE calling Next
	game1 := &types.GameParser{ID: 1}
	game2 := &types.GameParser{ID: 2}
	writeTestGameFile(t, tempDir, "game1.json", game1)
	writeTestGameFile(t, tempDir, "game2.json", game2)

	repo := NewGameParserRepository(tempDir)

	// First Next should return game1 immediately
	gotGame1, err := repo.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, game1.ID, gotGame1.ID)

	// Second Next should return game2 immediately (no blocking)
	gotGame2, err := repo.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, game2.ID, gotGame2.ID)

}

func TestGameParserRepository_Next_ContextCanceled(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewGameParserRepository(tempDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Next(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestGameParserRepository_Next_ReadDirError(t *testing.T) {
	// Pass an invalid directory path to trigger os.ReadDir error
	repo := NewGameParserRepository("/non/existing/directory")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	game, err := repo.Next(ctx)
	require.Nil(t, game)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no such file or directory")
}

func TestGameParserRepository_Next_ReadFileError(t *testing.T) {
	tempDir := t.TempDir()
	// Create a dummy JSON file but make it unreadable (permission denied)
	filePath := filepath.Join(tempDir, "game.json")
	err := os.WriteFile(filePath, []byte(`{}`), 0000) // no permissions
	require.NoError(t, err)

	repo := NewGameParserRepository(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	game, err := repo.Next(ctx)
	require.Nil(t, game)
	require.Error(t, err)
	// On some systems it may be permission denied
	require.Contains(t, err.Error(), "permission denied")
}

func TestGameParserRepository_Next_JSONUnmarshalError(t *testing.T) {
	tempDir := t.TempDir()

	// Write a malformed JSON file
	filePath := filepath.Join(tempDir, "game.json")
	err := os.WriteFile(filePath, []byte(`{ invalid json `), 0644)
	require.NoError(t, err)

	repo := NewGameParserRepository(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	game, err := repo.Next(ctx)
	require.Nil(t, game)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid character")
}

func TestGameParserRepository_Next_NoFilesWaits(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewGameParserRepository(tempDir) // empty dir, no JSON files

	ctx, cancel := context.WithTimeout(context.Background(), 1100*time.Millisecond)
	defer cancel()

	start := time.Now()
	game, err := repo.Next(ctx)
	duration := time.Since(start)

	require.Nil(t, game)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	// Check that it waited at least ~1 second (the context timeout)
	require.GreaterOrEqual(t, duration, 1*time.Second)
}
