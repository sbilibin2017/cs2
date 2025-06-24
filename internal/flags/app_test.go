package flags

import (
	"flag"
	"os"
	"testing"

	"github.com/sbilibin2017/cs2/internal/configs"
	"github.com/stretchr/testify/assert"
)

func reset() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	parserDir = ""
	dsn = ""
	logLevel = ""
}

func Test_parseAppFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected *configs.Config
	}{
		{
			name: "Default flags",
			args: []string{"cmd"},
			expected: &configs.Config{
				ParserDir:   "./data/raw",
				DatabaseDSN: "user:pass@localhost:5432/db",
				LogLevel:    "info",
			},
		},
		{
			name: "Custom parser dir",
			args: []string{"cmd", "-p", "/custom/parser"},
			expected: &configs.Config{
				ParserDir:   "/custom/parser",
				DatabaseDSN: "user:pass@localhost:5432/db",
				LogLevel:    "info",
			},
		},
		{
			name: "Custom all flags",
			args: []string{"cmd", "-p", "/data", "-d", "dsnstring", "-l", "debug"},
			expected: &configs.Config{
				ParserDir:   "/data",
				DatabaseDSN: "dsnstring",
				LogLevel:    "debug",
			},
		},
	}

	for _, tt := range tests {
		t.Run("parseAppFlags: "+tt.name, func(t *testing.T) {
			reset()
			os.Args = tt.args
			cfg := parseAppFlags()
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func Test_CombineAppFlags(t *testing.T) {
	reset()
	os.Args = []string{"cmd", "-p", "/combined", "-d", "combined_dsn", "-l", "warn"}
	cfg := ParseAppFlags()
	expected := &configs.Config{
		ParserDir:   "/combined",
		DatabaseDSN: "combined_dsn",
		LogLevel:    "warn",
	}
	assert.Equal(t, expected, cfg)
}
