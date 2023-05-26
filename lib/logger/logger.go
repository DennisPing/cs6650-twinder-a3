package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var zlog zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel) // Set default log level to INFO

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		level, err := zerolog.ParseLevel(logLevel)
		if err == nil {
			zerolog.SetGlobalLevel(level)
		}
	}

	fmt.Printf("Current log level: %s\n", zerolog.GlobalLevel().String())
	zlog = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func Info() *zerolog.Event {
	return zlog.Info()
}

func Error() *zerolog.Event {
	return zlog.Error()
}

func Warn() *zerolog.Event {
	return zlog.Warn()
}

func Debug() *zerolog.Event {
	return zlog.Debug()
}

func Fatal() *zerolog.Event {
	return zlog.Fatal()
}
