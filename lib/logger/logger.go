package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func InitLogger() zerolog.Logger {
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
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
