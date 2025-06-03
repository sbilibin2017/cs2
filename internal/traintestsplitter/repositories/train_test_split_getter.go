package repositories

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/traintestsplitter/types"
)

type TrainTestSplitGetterRepository struct {
	db clickhouse.Conn
}

func NewTrainTestSplitGetterRepository(db clickhouse.Conn) *TrainTestSplitGetterRepository {
	return &TrainTestSplitGetterRepository{db: db}
}

func (r *TrainTestSplitGetterRepository) Get(
	ctx context.Context,
) (*types.TrainTestSplit, error) {
	rows, err := r.db.Query(ctx, trainTestSplitQuery)
	if err != nil {
		logger.Log.Error(err)
		return nil, err
	}
	defer rows.Close()

	var split types.TrainTestSplit
	for rows.Next() {
		var trainIDs, testIDs []int32
		if err := rows.Scan(&split.Hash, &trainIDs, &testIDs); err != nil {
			logger.Log.Error(err)
			return nil, err
		}
		split.TrainGameIDs = trainIDs
		split.TestGameIDs = testIDs
	}

	return &split, nil
}

const trainTestSplitQuery = `
WITH
    (
        SELECT arrayReverse(arraySort(arrayDistinct(groupArray(game_id)))) AS sorted_game_ids
        FROM (
            SELECT game_id
            FROM games
            ORDER BY begin_at ASC
        )
    ) AS sorted_game_ids,
    arraySlice(sorted_game_ids, 1, 100) AS test_game_ids,
    arraySlice(sorted_game_ids, 101) AS train_game_ids,
    hex(MD5(arrayStringConcat(arrayMap(x -> toString(x), sorted_game_ids), ','))) AS hash
SELECT hash, train_game_ids, test_game_ids
`
