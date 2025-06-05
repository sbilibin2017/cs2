-- +goose Up
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_games_player_cumulative
ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(begin_at)
ORDER BY (player_id, begin_at, game_id)
POPULATE AS
SELECT
    player_id,
    game_id,
    begin_at,
    total_rounds,
    1 AS total_games,

    -- Per-game (current row) stats
    kills,
    deaths,
    assists,
    headshots,
    flash_assists,
    first_kills_diff,
    k_d_diff,
    adr,
    kast,
    rating,

    -- Total stats
    COUNT() OVER w AS total_games_cum,
    SUM(total_rounds) OVER w AS total_rounds_cum,

    SUM(kills) OVER w AS total_kills,
    SUM(deaths) OVER w AS total_deaths,
    SUM(assists) OVER w AS total_assists,
    SUM(headshots) OVER w AS total_headshots,
    SUM(flash_assists) OVER w AS total_flash_assists,

    SUM(first_kills_diff) OVER w AS total_first_kills_diff,
    SUM(k_d_diff) OVER w AS total_k_d_diff,
    SUM(adr) OVER w AS total_adr,          
    SUM(kast) OVER w AS total_kast,        
    SUM(rating) OVER w AS total_rating,    

    -- Derived total per-round stats
    total_kills / NULLIF(total_rounds_cum, 0) AS kills_per_round,
    total_deaths / NULLIF(total_rounds_cum, 0) AS deaths_per_round,
    total_assists / NULLIF(total_rounds_cum, 0) AS assists_per_round,
    total_headshots / NULLIF(total_rounds_cum, 0) AS headshots_per_round,
    total_flash_assists / NULLIF(total_rounds_cum, 0) AS flash_assists_per_round,

    -- Derived total per-game stats
    total_kills / NULLIF(total_games_cum, 0) AS kills_per_game,
    total_deaths / NULLIF(total_games_cum, 0) AS deaths_per_game,
    total_assists / NULLIF(total_games_cum, 0) AS assists_per_game,
    total_headshots / NULLIF(total_games_cum, 0) AS headshots_per_game,
    total_flash_assists / NULLIF(total_games_cum, 0) AS flash_assists_per_game,
    total_first_kills_diff / NULLIF(total_games_cum, 0) AS first_kills_diff_per_game,
    total_k_d_diff / NULLIF(total_games_cum, 0) AS kd_diff_per_game,
    total_adr / NULLIF(total_games_cum, 0) AS adr_per_game,
    total_kast / NULLIF(total_games_cum, 0) AS kast_per_game,
    total_rating / NULLIF(total_games_cum, 0) AS rating_per_game

FROM (
    SELECT
        player_id,
        game_id,
        begin_at, 
        MAX(round_id) AS total_rounds,

        AVG(kills) AS kills,
        AVG(deaths) AS deaths,
        AVG(assists) AS assists,
        AVG(headshots) AS headshots,
        AVG(flash_assists) AS flash_assists,

        AVG(first_kills_diff) AS first_kills_diff,
        AVG(k_d_diff) AS k_d_diff,
        AVG(adr) AS adr,
        AVG(kast) AS kast,
        AVG(rating) AS rating

    FROM games
    GROUP BY player_id, game_id, begin_at
)
WINDOW w AS (
    PARTITION BY player_id
    ORDER BY begin_at, game_id
    ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
);
- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS mv_games_player_cumulative;
-- +goose StatementEnd
