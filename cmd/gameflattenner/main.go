package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/sbilibin2017/cs2/internal/logger"
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
	flagGameRawDir     string
	flagGameFlattenDir string
	flagLogLevel       string
)

func parseFlags() {
	flag.StringVar(&flagGameRawDir, "r", "./data/raw", "Path to data directory")
	flag.StringVar(&flagGameFlattenDir, "f", "./data/flatten", "Path to data directory")
	flag.StringVar(&flagLogLevel, "l", "info", "Log level (e.g., debug, info, warn, error)")

	flag.Parse()
}

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go workers.StartGameFlattennerWorker(
		ctx,
		flagGameRawDir,
		flagGameFlattenDir,
	)

	<-ctx.Done()

	return nil
}
