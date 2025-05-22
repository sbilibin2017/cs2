package workers

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/sbilibin2017/cs2/internal/logger"
	"go.uber.org/zap"
)

type mapGame struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type league struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type serie struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Tier string `json:"tier"`
}

type tournament struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	PrizePool string `json:"prizepool"`
}

type match struct {
	League     league     `json:"league"`
	Serie      serie      `json:"serie"`
	Tournament tournament `json:"tournament"`
}

type team struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type player struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Hometown string `json:"hometown,omitempty"`
	Birthday string `json:"birthday,omitempty"`
}

type statistic struct {
	Team           team    `json:"team"`
	Player         player  `json:"player"`
	Kills          int     `json:"kills"`
	Deaths         int     `json:"deaths"`
	Assists        int     `json:"assists"`
	Headshots      int     `json:"headshots"`
	FlashAssists   int     `json:"flash_assists"`
	FirstKillsDiff int     `json:"first_kills_diff"`
	KDDiff         int     `json:"k_d_diff"`
	Adr            float32 `json:"adr"`
	Kast           float32 `json:"kast"`
	Rating         float32 `json:"rating"`
}

type round struct {
	Round      int    `json:"round"`
	Ct         int    `json:"ct"`
	Terrorists int    `json:"terrorists"`
	WinnerTeam int    `json:"winner_team"`
	Outcome    string `json:"outcome"`
}

type game struct {
	ID         int         `json:"id"`
	BeginAt    time.Time   `json:"begin_at"`
	Match      match       `json:"match"`
	Map        mapGame     `json:"map"`
	Statistics []statistic `json:"players"`
	Rounds     []round     `json:"rounds"`
}

type gameFlatten struct {
	ID               string    `json:"id"`
	BeginAt          time.Time `json:"begin_at"`
	GameID           int       `json:"game_id"`
	LeagueID         int       `json:"league_id"`
	SerieID          int       `json:"serie_id"`
	TournamentID     int       `json:"tournament_id"`
	MapID            int       `json:"map_id"`
	TeamID           int       `json:"team_id"`
	TeamOpponentID   int       `json:"team_opponent_id"`
	PlayerID         int       `json:"player_id"`
	PlayerOpponentID int       `json:"player_opponent_id"`
	RoundID          int       `json:"round_id"`
	Win              int       `json:"win"`
	Outcome          string    `json:"outcome"`
	Kills            int       `json:"kills"`
	Deaths           int       `json:"deaths"`
	Assists          int       `json:"assists"`
	Headshots        int       `json:"headshots"`
	FlashAssists     int       `json:"flash_assists"`
	FirstKillsDiff   int       `json:"first_kills_diff"`
	KDDiff           int       `json:"k_d_diff"`
	Adr              float32   `json:"adr"`
	Kast             float32   `json:"kast"`
	Rating           float32   `json:"rating"`
}

func StartGameParserWorker(
	ctx context.Context,
	pandascoreDirectory string,
	fileDirectory string,
	db clickhouse.Conn,
	interval int,
) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	gen := generatorNextFilePath(pandascoreDirectory)

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Game parser worker stopped due to context cancellation")
			return

		case <-ticker.C:
			filePath, err := gen()
			if err != nil {
				logger.Log.Error("Failed to get next file path", zap.Error(err))
				continue
			}
			if filePath == nil {
				logger.Log.Info("No file to process")
				continue
			}

			game, err := loadGameFromFile(*filePath)
			if err != nil {
				logger.Log.Error("Failed to load game from file", zap.String("path", *filePath), zap.Error(err))
				continue
			}

			if !isValidGameInfo(*game) {
				logger.Log.Error("Invalid game data, skipping", zap.Int("gameID", game.ID))
				continue
			}

			gamesFlatten := flattenGameStatistic(*game)
			logger.Log.Debug("Flattened game statistics",
				zap.Int("gameID", game.ID),
				zap.Int("records", len(gamesFlatten)),
			)

			err = insertFlattenGameStatistic(ctx, db, gamesFlatten)
			if err != nil {
				logger.Log.Error("Failed to insert flattened game stats",
					zap.Int("gameID", game.ID),
					zap.Error(err),
				)
				continue
			}

			logger.Log.Info("Successfully processed and inserted game stats",
				zap.Int("gameID", game.ID),
				zap.Int("records", len(gamesFlatten)),
			)
		}
	}
}

func generatorNextFilePath(
	filesDirectory string,
) func() (*string, error) {
	var files []string
	var currentIndex int

	return func() (*string, error) {
		if len(files) == 0 || currentIndex >= len(files) {
			entries, err := os.ReadDir(filesDirectory)
			if err != nil {
				return nil, err
			}

			files = nil
			for _, entry := range entries {
				if !entry.IsDir() {
					files = append(files, filepath.Join(filesDirectory, entry.Name()))
				}
			}

			if len(files) == 0 {
				return nil, nil
			}

			currentIndex = 0
		}

		file := files[currentIndex]
		currentIndex++
		return &file, nil
	}
}

func loadGameFromFile(filePath string) (*game, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var g game
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}

	return &g, nil
}

func isValidGameInfo(game game) bool {
	return isValidGame(game) &&
		isValidMap(game) &&
		isValidTeam(game) &&
		isValidPlayer(game) &&
		isValidRound(game)
}

func isValidGame(game game) bool {
	if game.ID == 0 || game.BeginAt.IsZero() {
		return false
	}
	return true
}

func isValidMap(game game) bool {
	if game.Map.ID == 0 ||
		game.Match.League.ID == 0 ||
		game.Match.Serie.ID == 0 ||
		game.Match.Tournament.ID == 0 {
		return false
	}
	return true
}

func isValidTeam(game game) bool {
	teamMap := make(map[int]int)
	for _, p := range game.Statistics {
		if p.Team.ID == 0 || p.Player.ID == 0 {
			return false
		}
		teamMap[p.Team.ID]++
	}
	return len(teamMap) == 2
}

func isValidPlayer(game game) bool {
	teamPlayers := map[int]map[int]struct{}{}
	for _, stat := range game.Statistics {
		if stat.Player.ID == 0 || stat.Team.ID == 0 {
			return false
		}
		if teamPlayers[stat.Team.ID] == nil {
			teamPlayers[stat.Team.ID] = make(map[int]struct{})
		}
		teamPlayers[stat.Team.ID][stat.Player.ID] = struct{}{}
	}
	if len(teamPlayers) != 2 {
		return false
	}
	for _, players := range teamPlayers {
		if len(players) != 5 {
			return false
		}
	}
	return true
}

func isValidRound(game game) bool {
	if len(game.Rounds) == 0 || game.Rounds[0].Round != 1 {
		return false
	}
	for _, r := range game.Rounds {
		if r.Round == 0 || r.Ct == 0 || r.Terrorists == 0 || r.WinnerTeam == 0 {
			return false
		}
	}
	lastRound := game.Rounds[len(game.Rounds)-1]
	return lastRound.Round >= 16
}

func flattenGameStatistic(game game) []gameFlatten {
	teams := make(map[int][]statistic)
	for _, stat := range game.Statistics {
		teams[stat.Team.ID] = append(teams[stat.Team.ID], stat)
	}

	var teamIDs []int
	for tid := range teams {
		teamIDs = append(teamIDs, tid)
	}
	sort.Ints(teamIDs)

	team1Players := teams[teamIDs[0]]
	team2Players := teams[teamIDs[1]]

	var result []gameFlatten

	for _, round := range game.Rounds {
		for _, p1 := range team1Players {
			for _, p2 := range team2Players {
				data := fmt.Sprintf("%d-%d-%d-%d-%d-%d-%d-%d-%d-%d",
					game.ID,
					game.Match.League.ID,
					game.Match.Serie.ID,
					game.Match.Tournament.ID,
					game.Map.ID,
					p1.Team.ID,
					p2.Team.ID,
					p1.Player.ID,
					p2.Player.ID,
					round.Round,
				)
				uuidStr := uuid.NewHash(sha1.New(), uuid.NameSpaceOID, []byte(data), 5).String()

				win := 0
				if p1.Team.ID == round.WinnerTeam {
					win = 1
				}

				flatten := gameFlatten{
					ID:               uuidStr,
					GameID:           game.ID,
					RoundID:          round.Round,
					LeagueID:         game.Match.League.ID,
					SerieID:          game.Match.Serie.ID,
					TournamentID:     game.Match.Tournament.ID,
					MapID:            game.Map.ID,
					TeamID:           p1.Team.ID,
					TeamOpponentID:   p2.Team.ID,
					PlayerID:         p1.Player.ID,
					PlayerOpponentID: p2.Player.ID,
					BeginAt:          game.BeginAt,
					Win:              win,
					Outcome:          round.Outcome,
					Kills:            p1.Kills,
					Deaths:           p1.Deaths,
					Assists:          p1.Assists,
					Headshots:        p1.Headshots,
					FlashAssists:     p1.FlashAssists,
					FirstKillsDiff:   p1.FirstKillsDiff,
					KDDiff:           p1.KDDiff,
					Adr:              p1.Adr,
					Kast:             p1.Kast,
					Rating:           p1.Rating,
				}

				result = append(result, flatten)
			}
		}
	}

	return result
}

func insertFlattenGameStatistic(
	ctx context.Context,
	db clickhouse.Conn,
	gamesFlatten []gameFlatten,
) error {
	const query = `
		INSERT INTO content.game_flatten (
			id, begin_at,
			game_id, 
			league_id, serie_id, tournament_id, 
			map_id,
			team_id, team_opponent_id, player_id, player_opponent_id,
			round_id, win, outcome, 
			kills, deaths, assists, headshots, flash_assists,
			first_kills_diff, k_d_diff, 
			adr, kast, rating
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	batch, err := db.PrepareBatch(ctx, query)
	if err != nil {
		return err
	}

	for _, gameFlatten := range gamesFlatten {
		err = batch.Append(
			gameFlatten.ID,
			gameFlatten.BeginAt,
			gameFlatten.GameID,
			gameFlatten.LeagueID,
			gameFlatten.SerieID,
			gameFlatten.TournamentID,
			gameFlatten.MapID,
			gameFlatten.TeamID,
			gameFlatten.TeamOpponentID,
			gameFlatten.PlayerID,
			gameFlatten.PlayerOpponentID,
			gameFlatten.RoundID,
			gameFlatten.Win,
			gameFlatten.Outcome,
			gameFlatten.Kills,
			gameFlatten.Deaths,
			gameFlatten.Assists,
			gameFlatten.Headshots,
			gameFlatten.FlashAssists,
			gameFlatten.FirstKillsDiff,
			gameFlatten.KDDiff,
			gameFlatten.Adr,
			gameFlatten.Kast,
			gameFlatten.Rating,
		)
		if err != nil {
			return err
		}
	}

	if err := batch.Send(); err != nil {
		return err
	}

	return nil
}
