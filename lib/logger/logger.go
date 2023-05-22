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

// StdLogger is a custom logger that implements the logger.Logger interface
// type StdLogger struct {
// 	zerolog.Logger
// }

// func (l StdLogger) Tracef(format string, args ...interface{}) {
// 	l.Trace().Msgf(format, args...)
// }

// func (l StdLogger) Debugf(format string, args ...interface{}) {
// 	l.Debug().Msgf(format, args...)
// }

// func (l StdLogger) Infof(format string, args ...interface{}) {
// 	l.Info().Msgf(format, args...)
// }

// func (l StdLogger) Printf(format string, args ...interface{}) {
// 	l.Info().Msgf(format, args...)
// }

// func (l StdLogger) Warnf(format string, args ...interface{}) {
// 	l.Warn().Msgf(format, args...)
// }

// func (l StdLogger) Errorf(format string, args ...interface{}) {
// 	l.Error().Msgf(format, args...)
// }

// func (l StdLogger) Fatalf(format string, args ...interface{}) {
// 	l.Fatal().Msgf(format, args...)
// }
