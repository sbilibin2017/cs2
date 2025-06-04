package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/sbilibin2017/cs2/internal/logger"
	"github.com/sbilibin2017/cs2/internal/repositories"
	"github.com/sbilibin2017/cs2/internal/services"
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
	flagGameValidDir           string
	flagDatasetFilePath        string
	flagTrainTestSplitFilePath string
	flagLableEncodingFilePath  string
	flagLogLevel               string
)

func parseFlags() {
	flag.StringVar(&flagGameValidDir, "v", "./data/valid", "Path to data directory")
	flag.StringVar(&flagTrainTestSplitFilePath, "s", "./data/train_test_split/train_test_split.json", "Path to data directory")
	flag.StringVar(&flagLableEncodingFilePath, "e", "./data/lable_encodings/player.json", "Path to data directory")
	flag.StringVar(&flagDatasetFilePath, "d", "./data/datasets/dataset.json", "Path to data directory")
	flag.StringVar(&flagLogLevel, "l", "info", "Log level (e.g., debug, info, warn, error)")

	flag.Parse()
}

func run() error {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		return err
	}

	lableEncoderRepository := repositories.NewLableEncoderRepository(
		flagGameValidDir,
	)
	trainTestSplitLoaderRepository := repositories.NewTrainTestSplitLoaderRepository(
		flagTrainTestSplitFilePath,
	)
	gameLoaderRepository := repositories.NewGameLoaderRepository(flagGameValidDir)
	datasetSaverRepository := repositories.NewDatasetSaverRepository(flagDatasetFilePath)

	datasetMakerService := services.NewDatasetMakerService(
		lableEncoderRepository,
		trainTestSplitLoaderRepository,
		gameLoaderRepository,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go workers.StartDatasetMakerWorker(
		ctx,
		datasetMakerService,
		datasetSaverRepository,
	)

	<-ctx.Done()

	return nil
}
