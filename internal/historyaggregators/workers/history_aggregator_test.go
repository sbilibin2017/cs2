package workers

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sbilibin2017/cs2/internal/historyaggregators/types"
	"github.com/stretchr/testify/require"
)

func TestConvertGameDBToFeatureExtractorParams_ExactlyFivePlayersEachTeam(t *testing.T) {
	now := time.Now()

	team1ID := int32(100)
	team2ID := int32(200)

	// Players for team 1 (unsorted)
	team1Players := []int32{10, 7, 5, 9, 8}
	// Players for team 2 (unsorted)
	team2Players := []int32{20, 18, 19, 16, 17}

	var games []types.GameDB

	// Create entries to simulate the players playing against each other
	roundID := int32(1)
	for i := 0; i < 5; i++ {
		games = append(games, types.GameDB{
			ID:               uuid.New(),
			BeginAt:          now,
			GameID:           42,
			LeagueID:         1,
			SerieID:          1,
			TournamentID:     1,
			TierID:           1,
			MapID:            1,
			TeamID:           team1ID,
			TeamOpponentID:   team2ID,
			PlayerID:         team1Players[i],
			PlayerOpponentID: team2Players[i],
			RoundID:          roundID,
			RoundIsCT:        1, // Set first round as CT for startCTTeamID
		})
		roundID++
	}

	// Run conversion
	params := convertGameDBToFeatureExtractorParams(games)

	// Check general fields
	require.Equal(t, now, params.BeginAt)
	require.Equal(t, int32(42), params.GameID)
	require.Equal(t, int32(1), params.LeagueID)

	// Teams sorted ascending
	require.Equal(t, [2]int32{team1ID, team2ID}, params.TeamIDs)

	// Players for team 1 sorted ascending (5 players)
	expectedTeam1 := []int32{5, 7, 8, 9, 10}
	require.Equal(t, expectedTeam1, params.PlayerIDs[:5])

	// Players for team 2 sorted ascending (5 players)
	expectedTeam2 := []int32{16, 17, 18, 19, 20}
	require.Equal(t, expectedTeam2, params.PlayerIDs[5:])

	// StartCTTeamID should be team1ID as RoundIsCT == 1 on first round
	require.Equal(t, team1ID, params.StartCTTeamID)
}
