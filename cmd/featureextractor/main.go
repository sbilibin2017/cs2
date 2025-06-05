package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os/signal"
	"syscall"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/repositories"
	"github.com/sbilibin2017/cs2/internal/workers"
)

func main() {
	parseFlags()
	err := run()
	if err != nil {
		panic(err)
	}
}

var (
	flagFeaturesDir string
	flagLogLevel    string
	flagDatabaseDSN string
)

func parseFlags() {
	flag.StringVar(&flagFeaturesDir, "s", "./data/features", "Path to data directory")
	flag.StringVar(&flagLogLevel, "l", "info", "Log level (e.g., debug, info, warn, error)")
	flag.StringVar(&flagDatabaseDSN, "d", "clickhouse://user:password@localhost:9000/db", "Database DSN (e.g., ClickHouse DSN)")

	flag.Parse()
}

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	logger.Log.Info("Parsing ClickHouse DSN...")
	opts, err := parseClickhouseDSN(flagDatabaseDSN)
	if err != nil {
		logger.Log.Errorf("Failed to parse DSN: %v", err)
		return err
	}
	logger.Log.Infof("Connecting to ClickHouse at %s...", opts.Addr)

	conn, err := clickhouse.Open(opts)
	if err != nil {
		logger.Log.Errorf("Failed to connect to ClickHouse: %v", err)
		return err
	}
	logger.Log.Info("Connected to ClickHouse")

	historyAggregatorPlayerRepository := repositories.NewHistoryAggregatorPlayerRepository(conn)
	featureExtractorPlayerRepository := repositories.NewFeatureExtractorPlayerRepository(historyAggregatorPlayerRepository)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go workers.StartFeatureExtractorWorker(
		ctx,
		flagFeaturesDir,
		conn,
		featureExtractorPlayerRepository,
	)

	<-ctx.Done()

	return nil
}

func parseClickhouseDSN(dsn string) (*clickhouse.Options, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %w", err)
	}

	host := u.Hostname()
	port := u.Port()
	addr := fmt.Sprintf("%s:%s", host, port)

	user := u.User.Username()
	password, _ := u.User.Password()

	database := u.Path
	if len(database) > 0 && database[0] == '/' {
		database = database[1:]
	}

	return &clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
	}, nil
}
