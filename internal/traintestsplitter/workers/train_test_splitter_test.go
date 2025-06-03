package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/cs2/internal/traintestsplitter/types"
)

func TestGeneratorTrainTestSplitParser_EmitsSplits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockTrainTestSplitGetter(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedSplit := &types.TrainTestSplit{
		Hash:         "abc123",
		TrainGameIDs: []int32{1, 2, 3},
		TestGameIDs:  []int32{4, 5},
	}

	callCount := 0
	mockGetter.EXPECT().Get(gomock.Any()).DoAndReturn(func(ctx context.Context) (*types.TrainTestSplit, error) {
		callCount++
		if callCount == 1 {
			return expectedSplit, nil
		}
		if callCount == 2 {
			return nil, nil
		}
		<-ctx.Done()
		return nil, ctx.Err()
	}).AnyTimes()

	ch := generatorTrainTestSplitParser(ctx, mockGetter, 10*time.Millisecond)

	select {
	case split := <-ch:
		assert.Equal(t, *expectedSplit, split)
	case <-time.After(1 * time.Second):
		t.Fatal("expected to receive a split but timed out")
	}
}

func TestGeneratorTrainTestSplitParser_HandlesGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockTrainTestSplitGetter(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	callCount := 0
	mockGetter.EXPECT().Get(gomock.Any()).DoAndReturn(func(ctx context.Context) (*types.TrainTestSplit, error) {
		callCount++
		switch callCount {
		case 1:
			return nil, errors.New("db error") // simulate error on first call
		case 2:
			cancel()        // cancel context to stop generator gracefully
			return nil, nil // simulate no data on second call (nil pointer)
		default:
			<-ctx.Done()
			return nil, ctx.Err()
		}
	}).AnyTimes()

	ch := generatorTrainTestSplitParser(ctx, mockGetter, 10*time.Millisecond)

	select {
	case split := <-ch:
		// If empty split struct received, treat as no output and return
		if split.Hash == "" && len(split.TrainGameIDs) == 0 && len(split.TestGameIDs) == 0 {
			return
		}
		t.Fatalf("expected no splits but got: %+v", split)
	case <-time.After(500 * time.Millisecond):
		// no output, test passes
	}
}

func TestSaverTrainTestSplit_SavesSplits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockTrainTestSplitSaver(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedSplit := types.TrainTestSplit{
		Hash:         "abc123",
		TrainGameIDs: []int32{1, 2, 3},
		TestGameIDs:  []int32{4, 5},
	}

	mockSaver.EXPECT().
		Save(gomock.Any(), gomock.AssignableToTypeOf(types.TrainTestSplit{})).
		DoAndReturn(func(ctx context.Context, split types.TrainTestSplit) error {
			if split.Hash != expectedSplit.Hash ||
				!equalInt32Slices(split.TrainGameIDs, expectedSplit.TrainGameIDs) ||
				!equalInt32Slices(split.TestGameIDs, expectedSplit.TestGameIDs) {
				t.Errorf("unexpected split passed to Save: got %+v want %+v", split, expectedSplit)
			}
			return nil
		}).Times(1)

	inCh := make(chan types.TrainTestSplit, 1)
	inCh <- expectedSplit
	close(inCh)

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	saverTrainTestSplit(ctx, mockSaver, inCh)

	// If no panic and no errors, test passed
}

func equalInt32Slices(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestStartParserWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockTrainTestSplitGetter(ctrl)
	mockSaver := NewMockTrainTestSplitSaver(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedSplit := &types.TrainTestSplit{
		Hash:         "hash123",
		TrainGameIDs: []int32{1, 2, 3},
		TestGameIDs:  []int32{4, 5},
	}

	callCount := 0
	mockGetter.EXPECT().Get(gomock.Any()).DoAndReturn(func(ctx context.Context) (*types.TrainTestSplit, error) {
		if callCount == 0 {
			callCount++
			return expectedSplit, nil
		}
		return nil, nil
	}).AnyTimes()

	mockSaver.EXPECT().Save(gomock.Any(), *expectedSplit).Return(nil).Times(1)

	go StartTrainTestSplitWorker(ctx, mockGetter, mockSaver, 10*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)
}
