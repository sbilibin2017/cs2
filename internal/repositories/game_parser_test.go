package repositories

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGameParserRepository_Next(t *testing.T) {
	ctx := context.Background()

	// Create temporary directory for test files
	dir := t.TempDir()

	// Prepare two minimal valid JSON game files
	gameJSON1 := `{
		"id": 1,
		"begin_at": "2023-01-01T00:00:00Z"
	}`
	gameJSON2 := `{
		"id": 2,
		"begin_at": "2023-02-02T00:00:00Z"
	}`

	err := os.WriteFile(filepath.Join(dir, "game1.json"), []byte(gameJSON1), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "game2.json"), []byte(gameJSON2), 0644)
	require.NoError(t, err)

	repo := NewGameParserRepository(WithPathToDir(dir)) // Use functional option

	// 1st call: should load files and return first game
	game, err := repo.Next(ctx)
	require.NoError(t, err)
	require.NotNil(t, game)
	require.Equal(t, int64(1), game.ID)

	// 2nd call: should return second game
	game, err = repo.Next(ctx)
	require.NoError(t, err)
	require.NotNil(t, game)
	require.Equal(t, int64(2), game.ID)

	// 3rd call: cycle back to first game (reload directory internally)
	game, err = repo.Next(ctx)
	require.NoError(t, err)
	require.NotNil(t, game)
	require.Equal(t, int64(1), game.ID)
}

func TestGameParserRepository_Next_EmptyDir(t *testing.T) {
	ctx := context.Background()

	emptyDir := t.TempDir()

	repo := NewGameParserRepository(WithPathToDir(emptyDir)) // Use functional option

	_, err := repo.Next(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no games found")
}

func TestGameParserRepository_Next_InvalidJSON(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "invalid.json"), []byte("not valid json"), 0644)
	require.NoError(t, err)

	repo := NewGameParserRepository(WithPathToDir(dir)) // Use functional option

	_, err = repo.Next(ctx)
	require.Error(t, err)
}
