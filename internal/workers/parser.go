package workers

import (
	"context"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
	"github.com/sbilibin2017/cs2/internal/validators"
)

type GameParser interface {
	Next(ctx context.Context) (*types.GameParser, error)
}

type GameSaver interface {
	Save(ctx context.Context, game types.GameParser) error
}

func StartParserWorker(
	ctx context.Context,
	p GameParser,
	s GameSaver,
) {
	gen := generatorGameParser(ctx, p)
	validCh := validateGameParser(ctx, gen)
	saveGame(ctx, s, validCh)
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

func saveGame(
	ctx context.Context,
	s GameSaver,
	gameCh <-chan types.GameParser,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case game, ok := <-gameCh:
			if !ok {
				continue
			}

			if err := s.Save(ctx, game); err != nil {
				logger.Log.Errorf("game_id=%d: %s", game.ID, err)
				continue
			}

			logger.Log.Infof("game_id=%d: ok", game.ID)
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
				if !validators.ValidateGame(g) {
					continue
				}
				validGamesChan <- g

			}
		}
	}()

	return validGamesChan
}
