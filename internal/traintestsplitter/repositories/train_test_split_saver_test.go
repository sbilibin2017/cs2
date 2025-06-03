package repositories_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sbilibin2017/cs2/internal/traintestsplitter/repositories"
	"github.com/sbilibin2017/cs2/internal/traintestsplitter/types"
	"github.com/stretchr/testify/require"
)

func TestTrainTestSplitSaverRepository_Save(t *testing.T) {
	// Setup temporary directory for test files
	tmpDir := t.TempDir()

	repo := repositories.NewTrainTestSplitSaverRepository(tmpDir)

	split := types.TrainTestSplit{
		Hash:         "testhash123",
		TrainGameIDs: []int32{1, 2, 3, 4},
		TestGameIDs:  []int32{100, 101},
	}

	err := repo.Save(context.Background(), split)
	require.NoError(t, err)

	// Check file exists
	filename := filepath.Join(tmpDir, split.Hash+".json")
	_, err = os.Stat(filename)
	require.NoError(t, err, "expected file to be created")

	// Read file and verify contents
	fileData, err := os.ReadFile(filename)
	require.NoError(t, err)

	var savedSplit types.TrainTestSplit
	err = json.Unmarshal(fileData, &savedSplit)
	require.NoError(t, err)

	require.Equal(t, split.Hash, savedSplit.Hash)
	require.Equal(t, split.TrainGameIDs, savedSplit.TrainGameIDs)
	require.Equal(t, split.TestGameIDs, savedSplit.TestGameIDs)
}
