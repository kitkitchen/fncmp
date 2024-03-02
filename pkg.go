package main

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
		Prefix:          "FnCmp: ",
	}),
}

type Config struct {
	DevMode  bool
	Silent   bool
	LogLevel LogLevel
	Logger   *log.Logger
}

func (c *Config) Set() {
	if c.Logger == nil {
		c.Logger = config.Logger
	}

	config = c
	c.Logger.Info("FnCmp config set")
	config.Logger.SetLevel(log.Level(c.LogLevel))
}
