package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/cs2/internal/types"
	"github.com/stretchr/testify/assert"
)

// --- Test NewParserWorker wrapper ---

func TestNewParserWorker_ReturnsFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParser := NewMockParser(ctrl)
	mockSaver := NewMockSaver(ctrl)

	// ✅ Use WithParser and WithSaver
	worker := NewParserWorker(
		WithParser(mockParser),
		WithSaver(mockSaver),
	)

	assert.NotNil(t, worker)
	assert.IsType(t, func(ctx context.Context) error { return nil }, worker)
}

// --- Test generatorGameParser ---

func TestGeneratorGameParser_ContextDone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParser := NewMockParser(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	ch := generatorGameParser(ctx, mockParser)

	// channel should be closed immediately
	_, ok := <-ch
	assert.False(t, ok)
}

func TestGeneratorGameParser_ReceivesGame(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParser := NewMockParser(ctrl)
	ctx := context.Background()

	expectedGame := &types.GameParser{ID: 123}
	callCount := 0
	mockParser.EXPECT().Next(gomock.Any()).DoAndReturn(func(ctx context.Context) (*types.GameParser, error) {
		callCount++
		if callCount == 1 {
			return expectedGame, nil
		}
		if callCount == 2 {
			return nil, nil // simulate no game but no error, will continue
		}
		return nil, errors.New("stop") // stop after 2 calls
	}).AnyTimes()

	ch := generatorGameParser(ctx, mockParser)

	received := <-ch
	assert.Equal(t, *expectedGame, received)
}

// --- Test flattenGameParser ---

func TestFlattenGameParser_ProperOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan types.GameParser, 1)

	game := types.GameParser{
		ID:      1,
		BeginAt: time.Now(),
		Match: types.MatchParser{
			League:     types.LeagueParser{ID: 10},
			Serie:      types.SerieParser{ID: 20, Tier: "a"},
			Tournament: types.TournamentParser{ID: 30},
		},
		Map: types.MapParser{ID: 100},
		Players: []types.PlayerStatisticParser{
			{Player: types.PlayerParser{ID: 1}, Team: types.TeamParser{ID: 1000}, Kills: 5, Deaths: 3, Assists: 1, Headshots: 2, FlashAssists: 0, KDDiff: 0.5, FirstKillsDiff: 1.0, ADR: 80.0, Kast: 0.6, Rating: 1.2},
			{Player: types.PlayerParser{ID: 6}, Team: types.TeamParser{ID: 2000}, Kills: 4, Deaths: 4, Assists: 2, Headshots: 1, FlashAssists: 1, KDDiff: 0.0, FirstKillsDiff: -0.5, ADR: 70.0, Kast: 0.5, Rating: 1.0},
		},
		Rounds: []types.RoundParser{
			{Round: 1, Outcome: "defused", WinnerTeam: 1000},
			{Round: 2, Outcome: "exploded", WinnerTeam: 2000},
		},
	}

	in <- game
	close(in)

	out := flattenGameParser(ctx, in)

	batch, ok := <-out
	assert.True(t, ok)
	assert.Greater(t, len(batch), 0)

	for _, g := range batch {
		assert.Equal(t, int64(game.ID), g.GameID)
		assert.Contains(t, []int64{1000, 2000}, g.TeamID)
		assert.Contains(t, []int64{1000, 2000}, g.TeamOpponentID)
	}

	_, ok = <-out
	assert.False(t, ok)
}

func TestFlattenGameParser_ContextCancelStops(t *testing.T) {
	in := make(chan types.GameParser)
	ctx, cancel := context.WithCancel(context.Background())

	out := flattenGameParser(ctx, in)

	cancel() // cancel context

	_, ok := <-out
	assert.False(t, ok)
}

// --- Test saveGameDB ---

func TestSaveGameDB_SavesBatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockSaver(ctrl)
	ctx := context.Background()

	batch := []types.GameDB{{GameID: 1}}
	in := make(chan []types.GameDB, 1)
	in <- batch
	close(in)

	mockSaver.EXPECT().Save(ctx, batch).Return(nil)

	errCh := saveGameDB(ctx, mockSaver, in)

	err, ok := <-errCh
	assert.False(t, ok) // channel closed without error
	assert.NoError(t, err)
}

func TestSaveGameDB_SaveErrorPropagated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockSaver(ctrl)
	ctx := context.Background()

	batch := []types.GameDB{{GameID: 1}}
	in := make(chan []types.GameDB, 1)
	in <- batch
	close(in)

	mockSaver.EXPECT().Save(ctx, batch).Return(errors.New("fail"))

	errCh := saveGameDB(ctx, mockSaver, in)

	err, ok := <-errCh
	assert.True(t, ok)
	assert.EqualError(t, err, "fail")

	// channel closes after sending error
	_, ok = <-errCh
	assert.False(t, ok)
}

func TestSaveGameDB_ContextCancelStops(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockSaver(ctrl)
	in := make(chan []types.GameDB)
	ctx, cancel := context.WithCancel(context.Background())

	errCh := saveGameDB(ctx, mockSaver, in)

	cancel()

	_, ok := <-errCh
	assert.False(t, ok)
}

// --- Test logErrors ---

func TestLogErrors_LogsErrorsAndReturnsNil(t *testing.T) {
	errCh := make(chan error, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh <- errors.New("some error")
	errCh <- nil
	close(errCh)

	err := logErrors(ctx, errCh)
	assert.NoError(t, err)
}

func TestParse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParser := NewMockParser(ctrl)
	mockSaver := NewMockSaver(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	game := &types.GameParser{
		ID:      1,
		BeginAt: time.Now(),
		Match: types.MatchParser{
			League:     types.LeagueParser{ID: 10},
			Serie:      types.SerieParser{ID: 20, Tier: "a"},
			Tournament: types.TournamentParser{ID: 30},
		},
		Map: types.MapParser{ID: 100},
		Players: []types.PlayerStatisticParser{
			{
				Player: types.PlayerParser{ID: 1},
				Team:   types.TeamParser{ID: 1000},
				Kills:  5, Deaths: 3, Assists: 1,
				Headshots: 2, FlashAssists: 0,
				KDDiff: 0.5, FirstKillsDiff: 1.0,
				ADR: 80.0, Kast: 0.6, Rating: 1.2,
			},
			{
				Player: types.PlayerParser{ID: 6},
				Team:   types.TeamParser{ID: 2000},
				Kills:  4, Deaths: 4, Assists: 2,
				Headshots: 1, FlashAssists: 1,
				KDDiff: 0.0, FirstKillsDiff: -0.5,
				ADR: 70.0, Kast: 0.5, Rating: 1.0,
			},
		},
		Rounds: []types.RoundParser{
			{Round: 1, Outcome: "defused", WinnerTeam: 1000},
		},
	}

	callCount := 0
	mockParser.EXPECT().
		Next(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (*types.GameParser, error) {
			if callCount == 0 {
				callCount++
				return game, nil
			}
			return nil, nil
		}).
		AnyTimes()

	mockSaver.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	err := parse(ctx, mockParser, mockSaver)
	assert.NoError(t, err)
}

func TestNewParserWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParser := NewMockParser(ctrl)
	mockSaver := NewMockSaver(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	game := &types.GameParser{
		ID:      1,
		BeginAt: time.Now(),
		Match: types.MatchParser{
			League:     types.LeagueParser{ID: 10},
			Serie:      types.SerieParser{ID: 20, Tier: "a"},
			Tournament: types.TournamentParser{ID: 30},
		},
		Map: types.MapParser{ID: 100},
		Players: []types.PlayerStatisticParser{
			{
				Player: types.PlayerParser{ID: 1},
				Team:   types.TeamParser{ID: 1000},
				Kills:  5, Deaths: 3, Assists: 1,
				Headshots: 2, FlashAssists: 0,
				KDDiff: 0.5, FirstKillsDiff: 1.0,
				ADR: 80.0, Kast: 0.6, Rating: 1.2,
			},
			{
				Player: types.PlayerParser{ID: 6},
				Team:   types.TeamParser{ID: 2000},
				Kills:  4, Deaths: 4, Assists: 2,
				Headshots: 1, FlashAssists: 1,
				KDDiff: 0.0, FirstKillsDiff: -0.5,
				ADR: 70.0, Kast: 0.5, Rating: 1.0,
			},
		},
		Rounds: []types.RoundParser{
			{Round: 1, Outcome: "defused", WinnerTeam: 1000},
		},
	}

	callCount := 0
	mockParser.EXPECT().
		Next(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (*types.GameParser, error) {
			if callCount == 0 {
				callCount++
				return game, nil
			}
			return nil, nil
		}).
		AnyTimes()

	mockSaver.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	workerFunc := NewParserWorker(
		WithParser(mockParser),
		WithSaver(mockSaver),
	)
	err := workerFunc(ctx)

	assert.NoError(t, err)
}
