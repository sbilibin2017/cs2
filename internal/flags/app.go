package flags

import (
	"flag"

	"github.com/sbilibin2017/cs2/internal/configs"
)

var (
	parserDir string
	dsn       string
	logLevel  string
)

func ParseAppFlags() *configs.Config {
	config := parseAppFlags()
	return config

}

func parseAppFlags() *configs.Config {
	flag.StringVar(&parserDir, "p", "./data/raw", "Directory for parser files")
	flag.StringVar(&dsn, "d", "user:pass@localhost:5432/db", "Database DSN")
	flag.StringVar(&logLevel, "l", "info", "Logging level (e.g. debug, info, warn, error)")

	flag.Parse()

	return configs.NewConfig(
		configs.WithParserDir(parserDir),
		configs.WithDatabaseDSN(dsn),
		configs.WithLogLevel(logLevel),
	)
}
