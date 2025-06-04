package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/sbilibin2017/cs2/internal/types"
	"github.com/stretchr/testify/require"
)

func TestTrainTestSplitterRepository_Split(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	// Prepare 150 games with incremental BeginAt timestamps
	const totalGames = 150
	for i := 1; i <= totalGames; i++ {
		gameMeta := types.GameMeta{
			ID:      int32(i),
			BeginAt: time.Now().Add(time.Duration(i) * time.Minute),
		}

		fpath := filepath.Join(dir, filepath.Join("", filepath.Base(filepath.Join("", strconv.Itoa(i)+".json"))))
		fpath = filepath.Join(dir, strconv.Itoa(i)+".json")

		f, err := os.Create(fpath)
		require.NoError(t, err)

		err = json.NewEncoder(f).Encode(gameMeta)
		require.NoError(t, err)
		f.Close()
	}

	repo := NewTrainTestSplitterRepository(dir)
	split, err := repo.Split(ctx)
	require.NoError(t, err)

	// Expect train size = totalGames - testSize (which is 100)
	require.Len(t, split.TrainIDs, totalGames-testSize)
	require.Len(t, split.TestIDs, testSize)

	// Check that IDs are sorted by BeginAt increasing (train first, then test)
	lastTrainID := int(0)
	for _, id := range split.TrainIDs {
		require.Greater(t, id, lastTrainID)
		lastTrainID = id
	}

	lastTestID := int(0)
	for _, id := range split.TestIDs {
		require.Greater(t, id, lastTestID)
		lastTestID = id
	}

	// Now test with fewer games than testSize, all go to test set
	t.Run("fewer than testSize", func(t *testing.T) {
		dir2 := t.TempDir()
		for i := 1; i <= 50; i++ {
			gameMeta := types.GameMeta{
				ID:      int32(i),
				BeginAt: time.Now().Add(time.Duration(i) * time.Minute),
			}
			fpath := filepath.Join(dir2, strconv.Itoa(i)+".json")
			f, err := os.Create(fpath)
			require.NoError(t, err)

			err = json.NewEncoder(f).Encode(gameMeta)
			require.NoError(t, err)
			f.Close()
		}

		repo2 := NewTrainTestSplitterRepository(dir2)
		split2, err := repo2.Split(ctx)
		require.NoError(t, err)

		require.Empty(t, split2.TrainIDs)
		require.Len(t, split2.TestIDs, 50)
	})
}
