package apps

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/configs"
	"github.com/sbilibin2017/cs2/internal/repositories"
	"github.com/sbilibin2017/cs2/internal/workers"
)

type App struct {
	config *configs.Config

	db clickhouse.Conn

	gameParserRepository *repositories.GameParserRepository
	gameSaverRepository  *repositories.GameSaverRepository

	Workers []func(ctx context.Context) error
}

func NewApp(config *configs.Config) (*App, error) {
	var app App
	app.config = config

	app.gameParserRepository = repositories.NewGameParserRepository(
		repositories.WithPathToDir(config.ParserDir),
	)
	app.gameSaverRepository = repositories.NewGameSaverRepository(
		repositories.WithDB(app.db),
	)

	app.Workers = []func(ctx context.Context) error{
		workers.NewParserWorker(
			workers.WithParser(app.gameParserRepository),
			workers.WithSaver(app.gameSaverRepository),
		),
	}

	return &app, nil
}

func (app *App) Run(ctx context.Context) error {
	opts, err := parseClickhouseDSN(app.config.DatabaseDSN)
	if err != nil {
		return err
	}

	app.db, err = clickhouse.Open(opts)
	if err != nil {
		return err
	}
	defer app.db.Close()

	if len(app.Workers) == 0 {
		return nil
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	var wg sync.WaitGroup
	errCh := make(chan error, len(app.Workers))

	for _, worker := range app.Workers {
		wg.Add(1)

		go func(w func(ctx context.Context) error) {
			defer wg.Done()

			if err := w(ctx); err != nil {
				errCh <- err
			}
		}(worker)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			log.Printf("worker error: %v", err)
		}
	}

	return nil
}

func parseClickhouseDSN(dsn string) (*clickhouse.Options, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %w", err)
	}

	if u.Scheme != "clickhouse" {
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		host += ":9000"
	}

	user := u.User.Username()
	password, _ := u.User.Password()

	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		database = "default"
	}

	return &clickhouse.Options{
		Addr: []string{host},
		Auth: clickhouse.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
	}, nil
}
