package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

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
	flagSource       string
	flagGameValidDir string
	flagLogLevel     string
)

func parseFlags() {
	flag.StringVar(&flagSource, "s", "./data/raw", "Path to data directory")
	flag.StringVar(&flagGameValidDir, "v", "./data/valid", "Path to data directory")
	flag.StringVar(&flagLogLevel, "l", "info", "Log level (e.g., debug, info, warn, error)")

	flag.Parse()
}

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	gameParserRepository := repositories.NewGameNextRepository(flagSource)
	gameSaverRepository := repositories.NewGameSaverRepository(flagGameValidDir)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go workers.StartParserWorker(
		ctx,
		gameParserRepository,
		gameSaverRepository,
	)

	<-ctx.Done()

	return nil
}
