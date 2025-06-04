package workers

import (
	"context"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

type TrainTestSplitter interface {
	Split(ctx context.Context) (*types.TrainTestSplit, error)
}

type TrainTestSplitSaver interface {
	Save(ctx context.Context, split types.TrainTestSplit) error
}

func StartTrainTestSplitterWorker(
	ctx context.Context,
	spiltter TrainTestSplitter,
	saver TrainTestSplitSaver,
) {
	gen := generatorTrainTestSplit(ctx, spiltter)
	saveSplit(ctx, saver, gen)
}

func generatorTrainTestSplit(
	ctx context.Context,
	s TrainTestSplitter,
) <-chan types.TrainTestSplit {
	out := make(chan types.TrainTestSplit, 100)

	go func() {
		defer func() {
			close(out)

		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				split, err := s.Split(ctx)
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

func saveSplit(
	ctx context.Context,
	saver TrainTestSplitSaver,
	splitCh <-chan types.TrainTestSplit,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case split, ok := <-splitCh:
			if !ok {
				continue
			}

			if err := saver.Save(ctx, split); err != nil {
				logger.Log.Errorf("split: %s", err)
				continue
			}

			logger.Log.Infof("split: ok")
		}
	}
}
