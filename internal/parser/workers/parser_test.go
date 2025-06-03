package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/cs2/internal/parser/types"
)

func createSampleGameParser() types.GameParser {
	return types.GameParser{
		ID:      1,
		BeginAt: time.Now(),
		Map: types.MapParser{
			ID:   1,
			Name: "de_dust2",
		},
		Match: types.MatchParser{
			League: types.LeagueParser{
				ID:   1,
				Name: "ESL",
			},
			Serie: types.SerieParser{
				ID:   1,
				Name: "Serie A",
				Tier: "s",
			},
			Tournament: types.TournamentParser{
				ID:        1,
				Name:      "Major",
				PrizePool: "1M$",
			},
		},
		Statistics: []types.StatisticParser{
			{Team: types.TeamParser{ID: 1}, Player: types.PlayerParser{ID: 101}, Kills: 10, Deaths: 5, Assists: 3, Headshots: 7, FlashAssists: 1, FirstKillsDiff: 2, KDDiff: 5, Adr: 85.5, Kast: 70.0, Rating: 1.2},
			{Team: types.TeamParser{ID: 1}, Player: types.PlayerParser{ID: 102}},
			{Team: types.TeamParser{ID: 1}, Player: types.PlayerParser{ID: 103}},
			{Team: types.TeamParser{ID: 1}, Player: types.PlayerParser{ID: 104}},
			{Team: types.TeamParser{ID: 1}, Player: types.PlayerParser{ID: 105}},

			{Team: types.TeamParser{ID: 2}, Player: types.PlayerParser{ID: 201}},
			{Team: types.TeamParser{ID: 2}, Player: types.PlayerParser{ID: 202}},
			{Team: types.TeamParser{ID: 2}, Player: types.PlayerParser{ID: 203}},
			{Team: types.TeamParser{ID: 2}, Player: types.PlayerParser{ID: 204}},
			{Team: types.TeamParser{ID: 2}, Player: types.PlayerParser{ID: 205}},
		},
		Rounds: []types.RoundParser{
			{Round: 1, Ct: 1, Terrorists: 2, WinnerTeam: 1, Outcome: "exploded"},
			{Round: 16, Ct: 2, Terrorists: 1, WinnerTeam: 2, Outcome: "defused"},
		},
	}
}

func TestStartParserWorker_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParser := NewMockGameParser(ctrl)
	mockSaver := NewMockGameSaver(ctrl)

	game := types.GameParser{
		ID: 1,
		// fill other necessary fields ...
	}

	called := false
	mockParser.EXPECT().
		Next(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (*types.GameParser, error) {
			if !called {
				called = true
				return &game, nil
			}
			return nil, nil
		}).
		AnyTimes()

	mockSaver.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	StartParserWorker(ctx, mockParser, mockSaver)
}

func TestStartParserWorker_ParserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParser := NewMockGameParser(ctrl)
	mockSaver := NewMockGameSaver(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Parser returns error repeatedly
	mockParser.EXPECT().
		Next(gomock.Any()).
		Return(nil, errors.New("parser error")).
		AnyTimes()

	// Saver Save should not be called at all
	mockSaver.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Times(0)

	go StartParserWorker(ctx, mockParser, mockSaver)

	<-ctx.Done()
}

func TestSaveGameDB_HandlesEmptyAndErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSaver := NewMockGameSaver(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Prepare GameDB slice
	records := []types.GameDB{
		{
			ID:               uuid.New(),
			BeginAt:          time.Now(),
			GameID:           1,
			RoundID:          1,
			RoundOutcomeID:   1,
			RoundIsCT:        1,
			LeagueID:         1,
			SerieID:          1,
			TournamentID:     1,
			TierID:           1,
			MapID:            1,
			TeamID:           1,
			TeamOpponentID:   2,
			PlayerID:         101,
			PlayerOpponentID: 201,
			Kills:            5,
			Deaths:           3,
			Assists:          2,
			Headshots:        1,
			FlashAssists:     0,
			FirstKillsDiff:   1,
			KDDiff:           2,
			Adr:              80.0,
			Kast:             75.0,
			Rating:           1.1,
			RoundWin:         1,
			UpdatedAt:        time.Now(),
		},
	}

	ch := make(chan []types.GameDB, 2)
	ch <- []types.GameDB{} // empty slice, should skip save
	ch <- records
	close(ch)

	// Saver Save returns error on call
	mockSaver.EXPECT().
		Save(gomock.Any(), records).
		Return(errors.New("save failed")).
		Times(1)

	// Run saveGameDB directly
	saveGameDB(ctx, mockSaver, ch)
}

func TestValidateGameParser_FiltersInvalidGames(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	invalidGame := types.GameParser{} // empty game, invalid
	validGame := createSampleGameParser()

	inputCh := make(chan types.GameParser, 2)
	inputCh <- invalidGame
	inputCh <- validGame
	close(inputCh)

	outputCh := validateGameParser(ctx, inputCh)

	var result []types.GameParser
	for g := range outputCh {
		result = append(result, g)
	}

	assert.Len(t, result, 1)
	assert.Equal(t, validGame.ID, result[0].ID)
}

func TestConvertGameParserToGameDB_ProducesRecords(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	game := createSampleGameParser()
	inputCh := make(chan types.GameParser, 1)
	inputCh <- game
	close(inputCh)

	outputCh := convertGameParserToGameDB(ctx, inputCh)

	records := <-outputCh
	assert.NotEmpty(t, records)
	for _, rec := range records {
		assert.Equal(t, game.ID, rec.GameID)
		assert.True(t, rec.ID != uuid.Nil)
	}
}
