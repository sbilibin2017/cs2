package workers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGeneratorNextFilePath_ReturnsFilesSequentially(t *testing.T) {
	dir := t.TempDir()

	// Создаём временные файлы
	files := []string{"a.json", "b.json", "c.json"}
	for _, name := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0644)
		require.NoError(t, err)
	}

	gen := generatorNextFilePath(dir)

	// Сортировка не гарантирована — сортируем ожидаемое
	expected := []string{
		filepath.Join(dir, "a.json"),
		filepath.Join(dir, "b.json"),
		filepath.Join(dir, "c.json"),
	}
	found := make(map[string]bool)

	// Получаем 3 разных пути
	for i := 0; i < len(files); i++ {
		p, err := gen()
		require.NoError(t, err)
		require.NotNil(t, p)
		found[*p] = true
	}

	// Проверяем, что все ожидаемые файлы были выданы
	for _, f := range expected {
		assert.True(t, found[f], "file %s was not returned by generator", f)
	}

	// После окончания — либо перескан, либо nil, в зависимости от поведения
	_, err := gen()
	assert.NoError(t, err)
}

func TestGeneratorNextFilePath_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	gen := generatorNextFilePath(dir)
	p, err := gen()
	require.NoError(t, err)
	assert.Nil(t, p)
}

func TestGeneratorNextFilePath_InvalidDirectory(t *testing.T) {
	invalidPath := "/path/does/not/exist"

	gen := generatorNextFilePath(invalidPath)
	p, err := gen()

	assert.Nil(t, p)
	assert.Error(t, err)
}

func TestLoadGameFromFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "game.json")

	content := `{
		"id": 1,
		"begin_at": "2025-01-01T12:00:00Z",
		"match": {
			"league": { "id": 10, "name": "Major League" },
			"serie": { "id": 20, "name": "Spring Split", "tier": "S" },
			"tournament": { "id": 30, "name": "Spring Finals", "prizepool": "$100k" }
		},
		"map": { "id": 5, "name": "Dust II" },
		"players": [
			{
				"team": { "id": 100, "name": "Team A", "location": "USA" },
				"player": { "id": 200, "name": "Player A" },
				"kills": 10,
				"deaths": 5,
				"assists": 3,
				"headshots": 2,
				"flash_assists": 1,
				"first_kills_diff": 1,
				"k_d_diff": 5,
				"adr": 95.5,
				"kast": 75.0,
				"rating": 1.2
			}
		],
		"rounds": [
			{ "round": 1, "ct": 5, "terrorists": 10, "winner_team": 100, "outcome": "bomb_defused" }
		]
	}`

	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0644))

	game, err := loadGameFromFile(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, game)

	assert.Equal(t, 1, game.ID)
	assert.Equal(t, "Major League", game.Match.League.Name)
	assert.Equal(t, "Dust II", game.Map.Name)
	assert.Equal(t, 1, len(game.Statistics))
	assert.Equal(t, 1, len(game.Rounds))
	assert.Equal(t, 100, game.Statistics[0].Team.ID)

	expectedTime, _ := time.Parse(time.RFC3339, "2025-01-01T12:00:00Z")
	assert.Equal(t, expectedTime, game.BeginAt)
}

func TestLoadGameFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad.json")

	badContent := `{"id": 1, "begin_at": "not-a-date",` // malformed JSON
	require.NoError(t, os.WriteFile(tmpFile, []byte(badContent), 0644))

	game, err := loadGameFromFile(tmpFile)
	assert.Nil(t, game)
	assert.Error(t, err)
}

func TestLoadGameFromFile_FileNotFound(t *testing.T) {
	game, err := loadGameFromFile("this_file_does_not_exist.json")
	assert.Nil(t, game)
	assert.Error(t, err)
}

func TestIsValidGame(t *testing.T) {
	valid := game{ID: 1, BeginAt: time.Now()}
	invalidID := game{ID: 0, BeginAt: time.Now()}
	invalidTime := game{ID: 1}

	assert.True(t, isValidGame(valid))
	assert.False(t, isValidGame(invalidID))
	assert.False(t, isValidGame(invalidTime))
}

func TestIsValidMap(t *testing.T) {
	valid := game{
		Map: mapGame{ID: 1},
		Match: match{
			League:     league{ID: 1},
			Serie:      serie{ID: 2},
			Tournament: tournament{ID: 3},
		},
	}
	missingMap := valid
	missingMap.Map.ID = 0
	missingLeague := valid
	missingLeague.Match.League.ID = 0

	assert.True(t, isValidMap(valid))
	assert.False(t, isValidMap(missingMap))
	assert.False(t, isValidMap(missingLeague))
}

func TestIsValidTeam(t *testing.T) {
	valid := game{
		Statistics: []statistic{
			{Team: team{ID: 1}, Player: player{ID: 10}},
			{Team: team{ID: 2}, Player: player{ID: 20}},
		},
	}
	oneTeam := game{
		Statistics: []statistic{
			{Team: team{ID: 1}, Player: player{ID: 10}},
			{Team: team{ID: 1}, Player: player{ID: 11}},
		},
	}
	invalidPlayer := game{
		Statistics: []statistic{
			{Team: team{ID: 1}, Player: player{ID: 0}},
			{Team: team{ID: 2}, Player: player{ID: 20}},
		},
	}

	assert.True(t, isValidTeam(valid))
	assert.False(t, isValidTeam(oneTeam))
	assert.False(t, isValidTeam(invalidPlayer))
}

func TestIsValidRound(t *testing.T) {
	valid := game{
		Rounds: []round{
			{Round: 1, Ct: 3, Terrorists: 5, WinnerTeam: 1, Outcome: "defuse"},
			{Round: 2, Ct: 5, Terrorists: 4, WinnerTeam: 2, Outcome: "time"},
			{Round: 16, Ct: 7, Terrorists: 8, WinnerTeam: 1, Outcome: "plant"},
		},
	}
	noRounds := game{}
	startsNotFromFirst := game{Rounds: []round{{Round: 2, Ct: 5, Terrorists: 5, WinnerTeam: 1}}}
	badData := game{Rounds: []round{{Round: 1, Ct: 0, Terrorists: 1, WinnerTeam: 1}}}
	shortGame := game{
		Rounds: []round{
			{Round: 1, Ct: 1, Terrorists: 1, WinnerTeam: 1},
			{Round: 15, Ct: 1, Terrorists: 1, WinnerTeam: 1},
		},
	}

	assert.True(t, isValidRound(valid))
	assert.False(t, isValidRound(noRounds))
	assert.False(t, isValidRound(startsNotFromFirst))
	assert.False(t, isValidRound(badData))
	assert.False(t, isValidRound(shortGame))
}

func TestIsValidPlayer(t *testing.T) {
	// Вспомогательная функция для создания statistic
	makeStat := func(teamID, playerID int) statistic {
		return statistic{
			Team:   team{ID: teamID},
			Player: player{ID: playerID},
		}
	}

	t.Run("valid game with 2 teams, 5 players each", func(t *testing.T) {
		game := game{}
		for teamID := 1; teamID <= 2; teamID++ {
			for playerID := 1; playerID <= 5; playerID++ {
				game.Statistics = append(game.Statistics, makeStat(teamID, playerID))
			}
		}
		assert.True(t, isValidPlayer(game))
	})

	t.Run("invalid - less than 5 players in one team", func(t *testing.T) {
		game := game{}
		// Первая команда 5 игроков
		for playerID := 1; playerID <= 5; playerID++ {
			game.Statistics = append(game.Statistics, makeStat(1, playerID))
		}
		// Вторая команда только 4 игрока
		for playerID := 1; playerID <= 4; playerID++ {
			game.Statistics = append(game.Statistics, makeStat(2, playerID))
		}
		assert.False(t, isValidPlayer(game))
	})

	t.Run("invalid - player ID is zero", func(t *testing.T) {
		game := game{
			Statistics: []statistic{
				makeStat(1, 0), // invalid player ID
				makeStat(2, 1),
				makeStat(2, 2),
				makeStat(2, 3),
				makeStat(2, 4),
				makeStat(2, 5),
			},
		}
		assert.False(t, isValidPlayer(game))
	})

	t.Run("invalid - team ID is zero", func(t *testing.T) {
		game := game{
			Statistics: []statistic{
				makeStat(0, 1), // invalid team ID
				makeStat(2, 1),
				makeStat(2, 2),
				makeStat(2, 3),
				makeStat(2, 4),
				makeStat(2, 5),
			},
		}
		assert.False(t, isValidPlayer(game))
	})

	t.Run("invalid - less than 2 teams", func(t *testing.T) {
		game := game{}
		for playerID := 1; playerID <= 5; playerID++ {
			game.Statistics = append(game.Statistics, makeStat(1, playerID))
		}
		assert.False(t, isValidPlayer(game))
	})

	t.Run("valid - duplicate player IDs in different teams", func(t *testing.T) {
		game := game{}
		for playerID := 1; playerID <= 5; playerID++ {
			game.Statistics = append(game.Statistics, makeStat(1, playerID))
			game.Statistics = append(game.Statistics, makeStat(2, playerID)) // same player IDs but different teams — valid
		}
		assert.True(t, isValidPlayer(game))
	})
}

func makeValidGame() game {
	// Создаем валидную структуру game, удовлетворяющую всем валидаторам
	g := game{
		ID:      1,
		BeginAt: time.Now(),
		Map:     mapGame{ID: 1},
		Match: match{
			League:     league{ID: 1},
			Serie:      serie{ID: 1},
			Tournament: tournament{ID: 1},
		},
		Rounds: []round{
			{Round: 1, Ct: 3, Terrorists: 4, WinnerTeam: 1, Outcome: "win"},
			{Round: 16, Ct: 5, Terrorists: 5, WinnerTeam: 2, Outcome: "win"},
		},
	}

	// Добавляем по 5 игроков в каждую из двух команд
	for teamID := 1; teamID <= 2; teamID++ {
		for playerID := 1; playerID <= 5; playerID++ {
			g.Statistics = append(g.Statistics, statistic{
				Team:   team{ID: teamID},
				Player: player{ID: playerID},
				Kills:  10,
				Deaths: 5,
			})
		}
	}

	return g
}

func TestIsValidGameInfo(t *testing.T) {
	validGame := makeValidGame()
	assert.True(t, isValidGameInfo(validGame), "valid game should be valid")

	t.Run("invalid game ID", func(t *testing.T) {
		g := makeValidGame()
		g.ID = 0
		assert.False(t, isValidGameInfo(g))
	})

	t.Run("invalid map ID", func(t *testing.T) {
		g := makeValidGame()
		g.Map.ID = 0
		assert.False(t, isValidGameInfo(g))
	})

	t.Run("invalid team count", func(t *testing.T) {
		g := makeValidGame()
		// Удалим игроков второй команды, останется только одна команда
		filtered := []statistic{}
		for _, stat := range g.Statistics {
			if stat.Team.ID == 1 {
				filtered = append(filtered, stat)
			}
		}
		g.Statistics = filtered
		assert.False(t, isValidGameInfo(g))
	})

	t.Run("invalid player count", func(t *testing.T) {
		g := makeValidGame()
		// Уменьшим количество игроков в одной из команд (4 вместо 5)
		filtered := []statistic{}
		count := 0
		for _, stat := range g.Statistics {
			if stat.Team.ID == 1 {
				if count < 4 {
					filtered = append(filtered, stat)
				}
				count++
			} else {
				filtered = append(filtered, stat)
			}
		}
		g.Statistics = filtered
		assert.False(t, isValidGameInfo(g))
	})

	t.Run("invalid rounds", func(t *testing.T) {
		g := makeValidGame()
		// Очистим раунды
		g.Rounds = []round{}
		assert.False(t, isValidGameInfo(g))
	})
}

func TestFlattenGameStatistic(t *testing.T) {
	game := game{
		ID:      123,
		BeginAt: time.Now(),
		Map:     mapGame{ID: 10},
		Match: match{
			League:     league{ID: 1},
			Serie:      serie{ID: 2},
			Tournament: tournament{ID: 3},
		},
		Statistics: []statistic{
			{
				Team:           team{ID: 1},
				Player:         player{ID: 100},
				Kills:          5,
				Deaths:         3,
				Assists:        2,
				Headshots:      1,
				FlashAssists:   0,
				FirstKillsDiff: 1,
				KDDiff:         2,
				Adr:            75.5,
				Kast:           65.3,
				Rating:         1.1,
			},
			{
				Team:           team{ID: 2},
				Player:         player{ID: 200},
				Kills:          7,
				Deaths:         4,
				Assists:        3,
				Headshots:      2,
				FlashAssists:   1,
				FirstKillsDiff: 0,
				KDDiff:         -1,
				Adr:            80.2,
				Kast:           70.0,
				Rating:         1.3,
			},
		},
		Rounds: []round{
			{
				Round:      1,
				Ct:         3,
				Terrorists: 2,
				WinnerTeam: 1,
				Outcome:    "win",
			},
		},
	}

	result := flattenGameStatistic(game)

	// Ожидаем: 1 раунд * 1 игрок из первой команды * 1 игрок из второй команды = 1 запись
	assert.Len(t, result, 1)

	f := result[0]

	assert.Equal(t, game.ID, f.GameID)
	assert.Equal(t, game.Map.ID, f.MapID)
	assert.Equal(t, 1, f.RoundID)
	assert.Equal(t, 1, f.TeamID)
	assert.Equal(t, 2, f.TeamOpponentID)
	assert.Equal(t, 100, f.PlayerID)
	assert.Equal(t, 200, f.PlayerOpponentID)
	assert.Equal(t, 1, f.Win) // Победила команда 1
	assert.Equal(t, "win", f.Outcome)
	assert.Equal(t, 5, f.Kills)
	assert.Equal(t, 3, f.Deaths)
	assert.Equal(t, 2, f.Assists)
	assert.Equal(t, 1, f.Headshots)
	assert.Equal(t, 0, f.FlashAssists)
	assert.Equal(t, 1, f.FirstKillsDiff)
	assert.Equal(t, 2, f.KDDiff)
	assert.InDelta(t, 75.5, f.Adr, 0.001)
	assert.InDelta(t, 65.3, f.Kast, 0.001)
	assert.InDelta(t, 1.1, f.Rating, 0.001)

	// Проверка что ID не пустое (UUID)
	assert.NotEmpty(t, f.ID)
}

func setupClickHouseContainer(ctx context.Context, t *testing.T) (clickhouse.Conn, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "clickhouse/clickhouse-server:latest",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"CLICKHOUSE_USER":     "default",
			"CLICKHOUSE_PASSWORD": "pass1234", // Задаём пароль
		},
		WaitingFor: wait.ForListeningPort("9000/tcp"),
	}

	chContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	host, err := chContainer.Host(ctx)
	assert.NoError(t, err)

	port, err := chContainer.MappedPort(ctx, "9000")
	assert.NoError(t, err)

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{host + ":" + port.Port()},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "pass1234", // Указываем тот же пароль
		},
		DialTimeout: 5 * time.Second,
	})
	assert.NoError(t, err)

	err = conn.Exec(ctx, "CREATE DATABASE IF NOT EXISTS content")
	assert.NoError(t, err)

	const createTableQuery = `
		CREATE TABLE IF NOT EXISTS content.game_flatten (
			id String,
			begin_at DateTime,
			game_id UInt64,
			league_id UInt64,
			serie_id UInt64,
			tournament_id UInt64,
			map_id UInt64,
			team_id UInt64,
			team_opponent_id UInt64,
			player_id UInt64,
			player_opponent_id UInt64,
			round_id UInt64,
			win UInt8,
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
		) ENGINE = Memory
	`

	err = conn.Exec(ctx, createTableQuery)
	assert.NoError(t, err)

	teardown := func() {
		chContainer.Terminate(ctx)
	}

	return conn, teardown
}

func TestInsertFlattenGameStatistic(t *testing.T) {
	ctx := context.Background()
	db, teardown := setupClickHouseContainer(ctx, t)
	defer teardown()

	gameFlats := []gameFlatten{
		{
			ID:               "uuid-test-1",
			BeginAt:          time.Now(),
			GameID:           1,
			LeagueID:         10,
			SerieID:          20,
			TournamentID:     30,
			MapID:            40,
			TeamID:           100,
			TeamOpponentID:   200,
			PlayerID:         1000,
			PlayerOpponentID: 2000,
			RoundID:          1,
			Win:              1,
			Outcome:          "win",
			Kills:            5,
			Deaths:           2,
			Assists:          3,
			Headshots:        1,
			FlashAssists:     0,
			FirstKillsDiff:   1,
			KDDiff:           2,
			Adr:              80.5,
			Kast:             75.0,
			Rating:           1.2,
		},
	}

	err := insertFlattenGameStatistic(ctx, db, gameFlats)
	assert.NoError(t, err)

	// Проверяем вставленные данные
	rows, err := db.Query(ctx, "SELECT id, game_id, team_id, player_id, round_id, win, outcome FROM content.game_flatten")
	assert.NoError(t, err)

	var (
		id       string
		gameID   uint64
		teamID   uint64
		playerID uint64
		roundID  uint64
		win      uint8
		outcome  string
	)
	found := false
	for rows.Next() {
		err = rows.Scan(&id, &gameID, &teamID, &playerID, &roundID, &win, &outcome)
		assert.NoError(t, err)
		if id == "uuid-test-1" {
			found = true
			assert.Equal(t, uint64(1), gameID)
			assert.Equal(t, uint64(100), teamID)
			assert.Equal(t, uint64(1000), playerID)
			assert.Equal(t, uint64(1), roundID)
			assert.Equal(t, uint8(1), win)
			assert.Equal(t, "win", outcome)
		}
	}
	assert.True(t, found, "Inserted record not found")
}

func TestStartGameParserWorker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go StartGameParserWorker(ctx, "", "", nil, 1)
	time.Sleep(3 * time.Second)

}
