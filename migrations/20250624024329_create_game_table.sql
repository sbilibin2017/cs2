-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS games (
    game_id Int64,
    begin_at DateTime,

    league_id Int64,
    serie_id Int64,
    tier_id Int64,
    tournament_id Int64,

    map_id Int64,

    team_id Int64,
    team_opponent_id Int64,
    player_id Int64,
    player_opponent_id Int64,

    kills Int64,
    deaths Int64,
    assists Int64,
    headshots Int64,
    flash_assists Int64,
    k_d_diff Float64,
    first_kills_diff Float64,
    adr Float64,
    kast Float64,
    rating Float64,

    round_id Int64,
    round_outcome_id Int64,
    round_win Int64
)
ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(begin_at)
ORDER BY (
    begin_at, 
    game_id,   
    league_id,
    serie_id,
    tier_id,
    tournament_id,
    map_id,
    round_id,
    team_id,
    team_opponent_id,
    player_id,
    player_opponent_id
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS games;

-- +goose StatementEnd
