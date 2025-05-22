-- +goose Up

-- создание базы
-- +goose StatementBegin
CREATE DATABASE IF NOT EXISTS content;
-- +goose StatementEnd

-- создание таблицы
-- +goose StatementBegin
CREATE TABLE content.game_flatten (
    id UUID,
    begin_at DateTime,
    game_id Int32,
    league_id Int32,
    serie_id Int32,
    tournament_id Int32,
    map_id Int32,
    team_id Int32,
    team_opponent_id Int32,
    player_id Int32,
    player_opponent_id Int32,
    round_id Int32,
    win Int8,
    outcome String,
    kills Int32,
    deaths Int32,
    assists Int32,
    headshots Int32,
    flash_assists Int32,
    first_kills_diff Int32,
    k_d_diff Int32,
    adr Float32,
    kast Float32,
    rating Float32
) 
ENGINE = MergeTree()
PARTITION BY toYYYYMM(begin_at)
ORDER BY (toYYYYMM(begin_at), begin_at, id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS content.game_flatten;
-- +goose StatementEnd
