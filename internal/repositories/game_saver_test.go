package repositories

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sbilibin2017/cs2/internal/types"
	"github.com/stretchr/testify/require"
)

func TestGameSaverRepository_Save(t *testing.T) {
	tmpDir := t.TempDir()

	repo := NewGameSaverRepository(tmpDir)

	game := types.GameParser{
		ID:      12345,
		BeginAt: time.Now(),
	}

	err := repo.Save(context.Background(), game)
	require.NoError(t, err)

	expectedFile := filepath.Join(tmpDir, "12345.json")
	_, err = os.Stat(expectedFile)
	require.NoError(t, err, "Expected file to be created")
}
