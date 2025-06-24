package configs

type Config struct {
	ParserDir   string
	DatabaseDSN string
	LogLevel    string
}

type Opt func(*Config)

func NewConfig(opts ...Opt) *Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func WithParserDir(dir string) Opt {
	return func(c *Config) {
		c.ParserDir = dir
	}
}

func WithDatabaseDSN(dsn string) Opt {
	return func(c *Config) {
		c.DatabaseDSN = dsn
	}
}

func WithLogLevel(level string) Opt {
	return func(c *Config) {
		c.LogLevel = level
	}
}
