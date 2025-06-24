package workers

import (
	"context"
	"log"

	"github.com/sbilibin2017/cs2/internal/types"
)

type Parser interface {
	Next(ctx context.Context) (*types.GameParser, error)
}

type Saver interface {
	Save(ctx context.Context, games []types.GameDB) error
}

type parserWorkerConfig struct {
	parser Parser
	saver  Saver
}

type ParserOpt func(*parserWorkerConfig)

func WithParser(p Parser) ParserOpt {
	return func(cfg *parserWorkerConfig) {
		cfg.parser = p
	}
}

func WithSaver(s Saver) ParserOpt {
	return func(cfg *parserWorkerConfig) {
		cfg.saver = s
	}
}

func NewParserWorker(opts ...ParserOpt) func(ctx context.Context) error {
	cfg := &parserWorkerConfig{}

	for _, opt := range opts {
		opt(cfg)
	}
	return func(ctx context.Context) error {
		return parse(ctx, cfg.parser, cfg.saver)
	}
}

func parse(
	ctx context.Context,
	p Parser,
	s Saver,
) error {
	genCh := generatorGameParser(ctx, p)
	flattenCh := flattenGameParser(ctx, genCh)
	errCh := saveGameDB(ctx, s, flattenCh)
	return logErrors(ctx, errCh)
}

func generatorGameParser(ctx context.Context, parser Parser) <-chan types.GameParser {
	ch := make(chan types.GameParser, 100)

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				game, err := parser.Next(ctx)
				if err != nil {
					return
				}
				if game == nil {
					continue
				}
				ch <- *game
			}
		}
	}()

	return ch
}

func flattenGameParser(ctx context.Context, in <-chan types.GameParser) <-chan []types.GameDB {
	out := make(chan []types.GameDB, 100)

	boolToInt64 := func(b bool) int64 {
		if b {
			return 1
		}
		return 0
	}

	serieTierMap := make(map[string]int)
	serieTierMap["s"] = 1
	serieTierMap["a"] = 2
	serieTierMap["b"] = 3
	serieTierMap["c"] = 4
	serieTierMap["d"] = 5

	roundOutcomeMap := make(map[string]int)
	roundOutcomeMap["exploded"] = 1
	roundOutcomeMap["defused"] = 2
	roundOutcomeMap["eliminated"] = 3
	roundOutcomeMap["timeout"] = 4

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case game, ok := <-in:
				if !ok {
					return
				}

				teamPlayers := make(map[int][]types.PlayerStatisticParser)
				for _, p := range game.Players {
					teamPlayers[int(p.Team.ID)] = append(teamPlayers[int(p.Team.ID)], p)
				}

				var teamIDs []int
				for tID := range teamPlayers {
					teamIDs = append(teamIDs, tID)
				}

				if len(teamIDs) != 2 {
					continue
				}

				teamPairIDs := [][]int{
					{teamIDs[0], teamIDs[1]},
					{teamIDs[1], teamIDs[0]},
				}

				tier, ok := serieTierMap[game.Match.Serie.Tier]
				if !ok {
					tier = 0
				}

				var batch []types.GameDB

				for _, pair := range teamPairIDs {
					tID, tOppID := pair[0], pair[1]
					ps := teamPlayers[tID]
					psOpp := teamPlayers[tOppID]

					for _, p := range ps {
						for _, pOpp := range psOpp {
							for _, r := range game.Rounds {
								outcomeID, ok := roundOutcomeMap[r.Outcome]
								if !ok {
									outcomeID = 0
								}
								gameDB := types.GameDB{
									GameID:  int64(game.ID),
									BeginAt: game.BeginAt,

									LeagueID:     int64(game.Match.League.ID),
									SerieID:      int64(game.Match.Serie.ID),
									TierID:       int64(tier),
									TournamentID: int64(game.Match.Tournament.ID),

									MapID: int64(game.Map.ID),

									TeamID:           int64(tID),
									TeamOpponentID:   int64(tOppID),
									PlayerID:         int64(p.Player.ID),
									PlayerOpponentID: int64(pOpp.Player.ID),

									Kills:          int64(p.Kills),
									Deaths:         int64(p.Deaths),
									Assists:        int64(p.Assists),
									Headshots:      int64(p.Headshots),
									FlashAssists:   int64(p.FlashAssists),
									KDDiff:         p.KDDiff,
									FirstKillsDiff: p.FirstKillsDiff,
									ADR:            p.ADR,
									Kast:           p.Kast,
									Rating:         p.Rating,

									RoundID:        int64(r.Round),
									RoundOutcomeID: int64(outcomeID),
									RoundWin:       boolToInt64(int(r.WinnerTeam) == tID),
								}

								batch = append(batch, gameDB)
							}
						}
					}
				}

				if len(batch) > 0 {
					out <- batch
				}
			}
		}
	}()

	return out
}

func saveGameDB(ctx context.Context, saver Saver, in <-chan []types.GameDB) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		for {
			select {
			case <-ctx.Done():
				return
			case batch, ok := <-in:
				if !ok {
					return
				}
				if err := saver.Save(ctx, batch); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	return errCh
}

func logErrors(ctx context.Context, in <-chan error) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-in:
				if !ok {
					return
				}
				if err != nil {
					log.Printf("error: %v", err)
				}
			}
		}
	}()
	return nil
}
