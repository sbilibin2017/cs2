package repositories

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLableEncodingSaverRepository_Save(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "encoding.json")

	repo := NewLableEncodingSaverRepository(filePath)

	encoding := map[int]int{
		10: 1,
		20: 2,
		30: 3,
	}

	err := repo.Save(context.Background(), encoding)
	assert.NoError(t, err, "Save should not return error")

	// Read back file
	data, err := os.ReadFile(filePath)
	assert.NoError(t, err, "should be able to read saved file")

	var loaded map[int]int
	err = json.Unmarshal(data, &loaded)
	assert.NoError(t, err, "should unmarshal saved JSON")

	assert.Equal(t, encoding, loaded, "saved and loaded maps should be equal")
}
