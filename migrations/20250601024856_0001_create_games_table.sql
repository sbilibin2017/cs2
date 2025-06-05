-- +goose Up
-- +goose StatementBegin
CREATE TABLE games (    
    game_id Int64,
    begin_at DateTime,

    league_id Int64,
    serie_id Int64,
    tournament_id Int64,
    tier_id Int64,

    map_id Int64,

    team_id Int64,
    team_opponent_id Int64,
    player_id Int64,
    player_opponent_id Int64,
    
    round_id Int64,
    round_outcome_id Int64,
    round_ct_id Int64,
    round_winner_id Int64,    

    kills Int64,
    deaths Int64,
    assists Int64,
    headshots Int64,
    flash_assists Int64,
    first_kills_diff Int64,
    k_d_diff Int64,
    adr Float64,
    kast Float64,
    rating Float64,
    
    updated_at DateTime
) 
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(begin_at)
ORDER BY (
    begin_at,
    game_id,
    round_id,
    team_id,
    team_opponent_id,
    player_id,
    player_opponent_id,
    league_id,
    serie_id,
    tournament_id,
    tier_id,
    map_id
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS games;
-- +goose StatementEnd