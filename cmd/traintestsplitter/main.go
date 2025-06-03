package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os/signal"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/traintestsplitter/repositories"
	"github.com/sbilibin2017/cs2/internal/traintestsplitter/workers"
)

func main() {
	parseFlags()
	err := run()
	if err != nil {
		panic(err)
	}
}

var (
	flagDestinationDir string
	flagDumpInterval   time.Duration
	flagLogLevel       string
	flagDatabaseDSN    string
)

func parseFlags() {
	flag.StringVar(&flagDestinationDir, "s", "./data/train_test_splits", "Path to train test splits directory")
	flag.DurationVar(&flagDumpInterval, "i", time.Duration(1*time.Second), "Dump split to file interval")
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

	db, err := clickhouse.Open(opts)
	if err != nil {
		logger.Log.Errorf("Failed to connect to ClickHouse: %v", err)
		return err
	}
	logger.Log.Info("Connected to ClickHouse")

	trainTestSplitGetter := repositories.NewTrainTestSplitGetterRepository(db)
	trainTestSplitSaverRepository := repositories.NewTrainTestSplitSaverRepository(flagDestinationDir)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Log.Info("Starting parser worker goroutine")
	go workers.StartTrainTestSplitWorker(
		ctx,
		trainTestSplitGetter,
		trainTestSplitSaverRepository,
		flagDumpInterval,
	)

	logger.Log.Info("Waiting for termination signal (SIGINT/SIGTERM)...")
	<-ctx.Done()

	logger.Log.Info("Termination signal received, shutting down")
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
