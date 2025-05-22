package main

import (
	"context"
	"fmt"
	"net/url"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/workers"
	"go.uber.org/zap"
)

func run() error {
	if err := logger.Initialize(logLevel); err != nil {
		logger.Log.Error("Logger is not initialized", zap.Error(err))
		return err
	}
	logger.Log.Info("Logger initialized", zap.String("level", logLevel))

	opts, err := parseClickHouseDSN(databaseDSN)
	if err != nil {
		logger.Log.Error("failed to parse clickhouse connection", zap.Error(err))
		return err
	}

	db, err := clickhouse.Open(opts)
	if err != nil {
		logger.Log.Error("failed to open clickhouse connection", zap.Error(err))
		return err
	}

	logger.Log.Info("Clickhouse connection established", zap.String("dsn", databaseDSN))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	workers.StartGameParserWorker(
		ctx,
		rawDirectory,
		flattenDirectory,
		db,
		interval,
	)

	<-ctx.Done()

	logger.Log.Info("Shutdown signal received, exiting")

	return nil
}

func parseClickHouseDSN(dsn string) (*clickhouse.Options, error) {
	const prefix = "clickhouse://"
	if !strings.HasPrefix(dsn, prefix) {
		return nil, fmt.Errorf("DSN must start with %s", prefix)
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	password, _ := u.User.Password()

	opts := &clickhouse.Options{
		Addr: []string{u.Host},
		Auth: clickhouse.Auth{
			Username: u.User.Username(),
			Password: password,
			Database: strings.TrimPrefix(u.Path, "/"),
		},
	}

	return opts, nil
}
