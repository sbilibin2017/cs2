package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/cs2/internal/traintestsplitter/types"
)

func TestTrainTestSplitGetterByHashRepository_GetLast(t *testing.T) {
	// Setup temp dir
	dir := t.TempDir()

	// Helper to write split JSON files
	writeSplit := func(filename string, updatedAt time.Time) {
		split := types.TrainTestSplit{
			UpdatedAt: updatedAt,
		}
		data, err := json.Marshal(split)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, filename), data, 0644)
		assert.NoError(t, err)
	}

	// Write 3 files with different UpdatedAt
	writeSplit("split1.json", time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	writeSplit("split2.json", time.Date(2025, 6, 3, 12, 0, 0, 0, time.UTC)) // latest
	writeSplit("split3.json", time.Date(2025, 5, 31, 0, 0, 0, 0, time.UTC))

	repo := NewTrainTestSplitGetterByHashRepository(dir)

	// Call GetLast
	split, err := repo.GetLast(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, split)
	assert.Equal(t, time.Date(2025, 6, 3, 12, 0, 0, 0, time.UTC), split.UpdatedAt)
}
