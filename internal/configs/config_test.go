package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		options  []Opt
		expected *Config
	}{
		{
			name:     "No options",
			options:  nil,
			expected: &Config{},
		},
		{
			name: "With ParserDir",
			options: []Opt{
				WithParserDir("/tmp/parsers"),
			},
			expected: &Config{
				ParserDir: "/tmp/parsers",
			},
		},
		{
			name: "With DatabaseDSN",
			options: []Opt{
				WithDatabaseDSN("user:pass@tcp(localhost:3306)/dbname"),
			},
			expected: &Config{
				DatabaseDSN: "user:pass@tcp(localhost:3306)/dbname",
			},
		},
		{
			name: "With LogLevel",
			options: []Opt{
				WithLogLevel("debug"),
			},
			expected: &Config{
				LogLevel: "debug",
			},
		},
		{
			name: "With All Options",
			options: []Opt{
				WithParserDir("/app/parsers"),
				WithDatabaseDSN("root@/mydb"),
				WithLogLevel("info"),
			},
			expected: &Config{
				ParserDir:   "/app/parsers",
				DatabaseDSN: "root@/mydb",
				LogLevel:    "info",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig(tt.options...)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
