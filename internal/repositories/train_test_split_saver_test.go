package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sbilibin2017/cs2/internal/types"
	"github.com/stretchr/testify/require"
)

func TestTrainTestSplitSaverRepository_Save(t *testing.T) {
	ctx := context.Background()

	// Create temp dir and file path
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "split.json")

	repo := NewTrainTestSplitSaverRepository(filePath)

	originalSplit := types.TrainTestSplit{
		TrainIDs: []int{1, 2, 3},
		TestIDs:  []int{4, 5},
	}

	err := repo.Save(ctx, originalSplit)
	require.NoError(t, err)

	// Read file back and decode
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	var loadedSplit types.TrainTestSplit
	err = json.NewDecoder(file).Decode(&loadedSplit)
	require.NoError(t, err)

	// Assert equality
	require.Equal(t, originalSplit.TrainIDs, loadedSplit.TrainIDs)
	require.Equal(t, originalSplit.TestIDs, loadedSplit.TestIDs)
}
