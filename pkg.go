// Package fncmp brings enhanced functionality to the Component interface.
//
// See: kitkitchen.github.io/docs/fncmp
package fncmp

import (
	"math"
	"os"
	"time"

	"github.com/charmbracelet/log"
)

var config *Config

var logOpts = log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	TimeFormat:      time.Kitchen,
	Prefix:          "package fncmp:",
}

type LogLevel log.Level

const (
	Debug LogLevel = -4
	Info  LogLevel = 0
	Warn  LogLevel = 4
	Error LogLevel = 8
	Fatal LogLevel = 12
	None  LogLevel = math.MaxInt32
)

func init() {
	config = &Config{
		CacheTimeOut: time.Minute * 30,
		LogLevel:     Error,
		Logger:       log.NewWithOptions(os.Stderr, logOpts),
	}
}

type Config struct {
	Silent       bool          // If true, no logs will be printed
	CacheTimeOut time.Duration // Default cache timeout
	LogLevel     LogLevel
	Logger       *log.Logger
}

func SetConfig(c *Config) {
	config = c
	config.Set()
}

func (c *Config) Set() {
	if c.Logger == nil {
		c.Logger = log.NewWithOptions(os.Stderr, logOpts)
	}

	config = c
	if c.Silent || c.LogLevel == 0 {
		c.Logger.SetLevel(log.Level(None))
		return
	}
	c.Logger.Info(
		"fncmp config set",
		"cache_timeout", c.CacheTimeOut,
		"log_level", c.LogLevel,
	)

	config.Logger.SetLevel(log.Level(c.LogLevel))
}
