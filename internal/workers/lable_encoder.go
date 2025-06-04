package workers

import (
	"context"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

type TrainTestSplitLoader interface {
	Load(ctx context.Context) (*types.TrainTestSplit, error)
}

type LableEncoder interface {
	Fit(ctx context.Context, gameIDs []int) error
	GetEncodingMap() map[int]int
}

type LableEncodingSaver interface {
	Save(ctx context.Context, encoding map[int]int) error
}

func StartLableEncoderWorker(
	ctx context.Context,
	sl TrainTestSplitLoader,
	e LableEncoder,
	s LableEncodingSaver,
) {
	gen := generatorLableEncoder(ctx, sl, e)
	saveEncoding(ctx, s, gen)
}

func generatorLableEncoder(
	ctx context.Context,
	sl TrainTestSplitLoader,
	e LableEncoder,
) <-chan map[int]int {
	out := make(chan map[int]int, 100)

	go func() {
		defer func() {
			close(out)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				split, err := sl.Load(ctx)

				if err != nil {
					logger.Log.Error(err)
					continue
				}

				err = e.Fit(ctx, split.TrainIDs)

				if err != nil {
					logger.Log.Error(err)
					continue
				}

				encoder := e.GetEncodingMap()

				if encoder == nil {
					continue
				}

				out <- encoder
			}
		}
	}()

	return out
}

func saveEncoding(
	ctx context.Context,
	s LableEncodingSaver,
	encodingCh <-chan map[int]int,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case encoding, ok := <-encodingCh:
			if !ok {
				continue
			}

			if err := s.Save(ctx, encoding); err != nil {
				logger.Log.Errorf("error: %s", err)
				continue
			}

			logger.Log.Info("ok")
		}
	}
}
