package fncmp

import (
	"math"
	"os"
	"time"

	"github.com/charmbracelet/log"
)

type LogLevel log.Level

const (
	Debug LogLevel = -4
	Info  LogLevel = 0
	Warn  LogLevel = 4
	Error LogLevel = 8
	Fatal LogLevel = 12
	None  LogLevel = math.MaxInt32
)

var config = &Config{
	Silent:   false,
	LogLevel: Error,
	DevMode:  false,
	Logger: log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "package fncmp:",
	}),
}

type Config struct {
	DevMode  bool
	Silent   bool
	LogLevel LogLevel
	Logger   *log.Logger
}

func SetConfig(c *Config) {
	config = c
	config.Set()
}

func (c *Config) Set() {
	if c.Logger == nil {
		c.Logger = config.Logger
	}

	config = c
	if c.Silent || c.LogLevel == 0 {
		config.Logger.SetLevel(log.Level(None))
		return
	}
	c.Logger.Info("fncmp config set", "silent", c.Silent, "log_level", c.LogLevel)
	config.Logger.SetLevel(log.Level(c.LogLevel))
}
