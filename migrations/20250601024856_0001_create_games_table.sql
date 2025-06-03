-- +goose Up
-- +goose StatementBegin
CREATE TABLE games (
    id UUID,
    begin_at DateTime,

    game_id Int32,    

    league_id Int32,
    serie_id Int32,
    tournament_id Int32,
    tier_id Int32,
    map_id Int32,

    team_id Int32,
    team_opponent_id Int32,
    player_id Int32,
    player_opponent_id Int32,

    kills Int32,
    deaths Int32,
    assists Int32,
    headshots Int32,
    flash_assists Int32,
    first_kills_diff Int32,
    k_d_diff Int32,
    adr Float32,
    kast Float32,
    rating Float32,

    round_id Int32,
    round_outcome_id Int32,
    round_is_ct Int32,
    round_win Int32,
    
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
