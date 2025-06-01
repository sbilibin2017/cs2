package workers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

type GameParser interface {
	Next(ctx context.Context) (*types.GameParser, error)
}

type GameSaver interface {
	Save(ctx context.Context, games []types.GameDB) error
}

func StartParserWorker(
	ctx context.Context,
	p GameParser,
	s GameSaver,
) {
	gen := generatorGameParser(ctx, p)
	validCh := validateGameParser(ctx, gen)
	dbCh := convertGameParserToGameDB(ctx, validCh)
	saveGameDB(ctx, s, dbCh)
}

func generatorGameParser(
	ctx context.Context,
	p GameParser,
) <-chan types.GameParser {
	out := make(chan types.GameParser, 100)

	go func() {
		defer func() {
			close(out)

		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				game, err := p.Next(ctx)
				if err != nil {
					logger.Log.Error(err)
					continue
				}
				if game == nil {
					continue
				}

				out <- *game
			}
		}
	}()

	return out
}

func saveGameDB(
	ctx context.Context,
	s GameSaver,
	gameDBChan <-chan []types.GameDB,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case gameDBRecords, ok := <-gameDBChan:
			if !ok {
				continue
			}

			if len(gameDBRecords) == 0 {
				continue
			}

			if err := s.Save(ctx, gameDBRecords); err != nil {
				logger.Log.Error(err)
				continue
			}

			logger.Log.Infof("Saved game ID %d with %d records", gameDBRecords[0].GameID, len(gameDBRecords))
		}
	}
}

func validateGameParser(ctx context.Context, gameParserChan <-chan types.GameParser) chan types.GameParser {
	validGamesChan := make(chan types.GameParser, 100)

	go func() {
		defer close(validGamesChan)

		for {
			select {
			case <-ctx.Done():
				return
			case g, ok := <-gameParserChan:
				if !ok {
					return
				}

				if g.ID == 0 || g.BeginAt.IsZero() {
					continue
				}
				if g.Map.ID == 0 {
					continue
				}
				if g.Match.League.ID == 0 || g.Match.Serie.ID == 0 || g.Match.Tournament.ID == 0 {
					continue
				}

				teams := map[int32]struct{}{}
				invalid := false
				for _, stat := range g.Statistics {
					if stat.Team.ID == 0 {
						invalid = true
						break
					}
					teams[stat.Team.ID] = struct{}{}
				}
				if invalid || len(teams) != 2 {
					continue
				}

				players := map[int32]struct{}{}
				for _, stat := range g.Statistics {
					if stat.Player.ID == 0 {
						invalid = true
						break
					}
					players[stat.Player.ID] = struct{}{}
				}
				if invalid || len(players) != 10 {
					continue
				}

				if len(g.Rounds) == 0 || g.Rounds[0].Round != 1 {
					continue
				}

				for _, r := range g.Rounds {
					if r.Round == 0 || r.Ct == 0 || r.Terrorists == 0 || r.WinnerTeam == 0 {
						invalid = true
						break
					}
				}
				if invalid {
					continue
				}

				lastRound := g.Rounds[len(g.Rounds)-1]
				if lastRound.Round < 16 {
					continue
				}

				validGamesChan <- g

			}
		}
	}()

	return validGamesChan
}

func convertGameParserToGameDB(ctx context.Context, gameParserChan <-chan types.GameParser) chan []types.GameDB {
	resultChan := make(chan []types.GameDB, 100)

	go func() {
		defer close(resultChan)

		for {
			select {
			case <-ctx.Done():
				return
			case game, ok := <-gameParserChan:
				if !ok {
					return
				}

				teamPlayerIDMap := make(map[int32][]int32)
				teamOpponentMapID := make(map[int32]int32)

				for _, stat := range game.Statistics {
					teamPlayerIDMap[stat.Team.ID] = append(teamPlayerIDMap[stat.Team.ID], stat.Player.ID)
				}

				var teams []int32
				for teamID := range teamPlayerIDMap {
					teams = append(teams, teamID)
				}
				if len(teams) == 2 {
					teamOpponentMapID[teams[0]] = teams[1]
					teamOpponentMapID[teams[1]] = teams[0]
				}

				playerStatisticMap := make(map[int32]types.StatisticParser)
				for _, stat := range game.Statistics {
					playerStatisticMap[stat.Player.ID] = stat
				}

				tierMap := map[string]int32{"s": 1, "a": 2, "b": 3, "c": 4, "d": 5}
				tierID, ok := tierMap[game.Match.Serie.Tier]
				if !ok {
					tierID = 0
				}

				outcomeMap := map[string]int32{"exploded": 1, "defused": 2, "timeout": 3, "eliminated": 4}

				var result []types.GameDB

				teamPair1 := []int32{teams[0], teams[1]}
				teamPair2 := []int32{teams[1], teams[0]}
				teamPairs := [][]int32{teamPair1, teamPair2}

				for _, round := range game.Rounds {
					for _, teams := range teamPairs {
						for _, teamID := range teams {
							opponentTeamID := teamOpponentMapID[teamID]
							playersID := teamPlayerIDMap[teamID]
							playersOppID := teamPlayerIDMap[opponentTeamID]
							for _, playerID := range playersID {
								stat, ok := playerStatisticMap[playerID]
								if !ok {
									continue
								}

								win := int32(0)
								if round.WinnerTeam == teamID {
									win = 1
								}

								isCT := int32(0)
								if round.Ct == teamID {
									isCT = 1
								}

								outcomeID, ok := outcomeMap[round.Outcome]
								if !ok {
									outcomeID = 0
								}

								for _, playerOppID := range playersOppID {
									dbRecord := types.GameDB{
										ID:               uuid.New(),
										BeginAt:          game.BeginAt,
										GameID:           game.ID,
										RoundID:          round.Round,
										RoundOutcomeID:   outcomeID,
										RoundIsCT:        isCT,
										LeagueID:         game.Match.League.ID,
										SerieID:          game.Match.Serie.ID,
										TournamentID:     game.Match.Tournament.ID,
										TierID:           tierID,
										MapID:            game.Map.ID,
										TeamID:           teamID,
										TeamOpponentID:   opponentTeamID,
										PlayerID:         playerID,
										PlayerOpponentID: playerOppID,

										Kills:          stat.Kills,
										Deaths:         stat.Deaths,
										Assists:        stat.Assists,
										Headshots:      stat.Headshots,
										FlashAssists:   stat.FlashAssists,
										FirstKillsDiff: stat.FirstKillsDiff,
										KDDiff:         stat.KDDiff,
										Adr:            stat.Adr,
										Kast:           stat.Kast,
										Rating:         stat.Rating,

										Win:       win,
										UpdatedAt: time.Now().UTC(),
									}

									result = append(result, dbRecord)
								}
							}
						}
					}
				}

				resultChan <- result
			}
		}
	}()

	return resultChan
}
