package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	rawDirectory     string
	flattenDirectory string
	databaseDSN      string
	logLevel         string
	interval         int
)

func parseFlags() {
	flag.StringVar(&rawDirectory, "r", "./data/raw", "Path to source directory for game files")
	flag.StringVar(&flattenDirectory, "f", "./data/flatten", "Path to directory for flatten game files")
	flag.IntVar(&interval, "i", 1, "Interval for reading source files in seconds")
	flag.StringVar(&databaseDSN, "d", "", "Database DSN")
	flag.StringVar(&logLevel, "l", "", "Logging level (e.g. debug, info, warn, error)")

	flag.Parse()

	if env := os.Getenv("RAW_DIRECTORY"); env != "" {
		rawDirectory = env
	}

	if env := os.Getenv("FLATTEN_DIRECTORY"); env != "" {
		flattenDirectory = env
	}

	if env := os.Getenv("INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			interval = v
		}
	}

	if env := os.Getenv("DATABASE_DSN"); env != "" {
		databaseDSN = env
	}

	if env := os.Getenv("LOG_LEVEL"); env != "" {
		logLevel = env
	}

}
