package workers

import (
	"context"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/types"
)

type DatasetMaker interface {
	MakeDataset(ctx context.Context) ([]types.DatasetRow, error)
}

type DatasetSaver interface {
	Save(ctx context.Context, dataset []types.DatasetRow) error
}

func StartDatasetMakerWorker(
	ctx context.Context,
	m DatasetMaker,
	s DatasetSaver,
) {
	gen := generatorDataset(ctx, m)
	saveDataset(ctx, s, gen)
}

func generatorDataset(
	ctx context.Context,
	m DatasetMaker,
) <-chan []types.DatasetRow {
	out := make(chan []types.DatasetRow, 100)

	go func() {
		defer func() {
			close(out)

		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				rows, err := m.MakeDataset(ctx)
				if err != nil {
					logger.Log.Error(err)
					continue
				}
				if len(rows) == 0 {
					continue
				}

				out <- rows
			}
		}
	}()

	return out
}

func saveDataset(
	ctx context.Context,
	s DatasetSaver,
	in <-chan []types.DatasetRow,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case row, ok := <-in:
			if !ok {
				continue
			}

			if err := s.Save(ctx, row); err != nil {
				logger.Log.Errorf("split: %s", err)
				continue
			}

			logger.Log.Infof("split: ok")
		}
	}
}
