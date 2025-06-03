package workers

import (
	"context"
	"time"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/traintestsplitter/types"
)

type TrainTestSplitGetter interface {
	Get(ctx context.Context) (*types.TrainTestSplit, error)
}

type TrainTestSplitSaver interface {
	Save(ctx context.Context, split types.TrainTestSplit) error
}

func StartTrainTestSplitWorker(
	ctx context.Context,
	g TrainTestSplitGetter,
	s TrainTestSplitSaver,
	i time.Duration,
) {
	gen := generatorTrainTestSplitParser(ctx, g, i)
	saverTrainTestSplit(ctx, s, gen)
}

func generatorTrainTestSplitParser(
	ctx context.Context,
	g TrainTestSplitGetter,
	interval time.Duration,
) <-chan types.TrainTestSplit {
	out := make(chan types.TrainTestSplit, 100)

	go func() {
		defer close(out)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				split, err := g.Get(ctx)
				if err != nil {
					logger.Log.Error(err)
					continue
				}
				if split == nil {
					continue
				}
				out <- *split
			}
		}
	}()

	return out
}

func saverTrainTestSplit(
	ctx context.Context,
	s TrainTestSplitSaver,
	in <-chan types.TrainTestSplit,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}
			if err := s.Save(ctx, item); err != nil {
				logger.Log.Error(err)
				continue
			}

			logger.Log.Infof("Saved train test split: hash=%s", item.Hash)
		}
	}
}
