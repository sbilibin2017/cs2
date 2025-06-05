package repositories

import (
	"context"
	"time"

	"github.com/sbilibin2017/cs2/internal/types"
)

type HistoryAgregator interface {
	Aggregate(
		ctx context.Context,
		beginAt time.Time,
		playerID int64,
	) (*types.AggregationResult, error)
}

type FeatureExtractorPlayerRepository struct {
	ha HistoryAgregator
}

func NewFeatureExtractorPlayerRepository(ha HistoryAgregator) *FeatureExtractorPlayerRepository {
	return &FeatureExtractorPlayerRepository{
		ha: ha,
	}
}

func (r *FeatureExtractorPlayerRepository) Extract(
	ctx context.Context,
	params types.FeaturePlayerParams,
) (map[string]types.FeaturePlayer, error) {
	features := make(map[string][]float64)

	for _, playerID := range params.PlayerIDs {
		agg, err := r.ha.Aggregate(ctx, params.BeginAt, playerID)
		if err != nil {
			return nil, err
		}

		if agg == nil {
			continue
		}

		features["kills_total"] = append(features["kills_total"], agg.KillsTotal)
		features["deaths_total"] = append(features["deaths_total"], agg.DeathsTotal)
		features["assists_total"] = append(features["assists_total"], agg.AssistsTotal)
		features["headshots_total"] = append(features["headshots_total"], agg.HeadshotsTotal)
		features["flash_assists_total"] = append(features["flash_assists_total"], agg.FlashAssistsTotal)

		features["kills_per_round"] = append(features["kills_per_round"], agg.KillsPerRound)
		features["deaths_per_round"] = append(features["deaths_per_round"], agg.DeathsPerRound)
		features["assists_per_round"] = append(features["assists_per_round"], agg.AssistsPerRound)
		features["headshots_per_round"] = append(features["headshots_per_round"], agg.HeadshotsPerRound)
		features["flash_assists_per_round"] = append(features["flash_assists_per_round"], agg.FlashAssistsPerRound)

		features["kills_per_game"] = append(features["kills_per_game"], agg.KillsPerGame)
		features["deaths_per_game"] = append(features["deaths_per_game"], agg.DeathsPerGame)
		features["assists_per_game"] = append(features["assists_per_game"], agg.AssistsPerGame)
		features["headshots_per_game"] = append(features["headshots_per_game"], agg.HeadshotsPerGame)
		features["flash_assists_per_game"] = append(features["flash_assists_per_game"], agg.FlashAssistsPerGame)

		features["first_kills_diff_per_game"] = append(features["first_kills_diff_per_game"], agg.FirstKillsDiffPerGame)
		features["kd_diff_per_game"] = append(features["kd_diff_per_game"], agg.KDdiffPerGame)
		features["adr_per_game"] = append(features["adr_per_game"], agg.ADRPerGame)
		features["kast_per_game"] = append(features["kast_per_game"], agg.KASTPerGame)
		features["rating_per_game"] = append(features["rating_per_game"], agg.RatingPerGame)
	}

	featruesFinal := make(map[string]types.FeaturePlayer)

	for k, v := range features {
		featruesFinal[k] = *types.NewFeaturePlayer(v)

	}

	return featruesFinal, nil
}
