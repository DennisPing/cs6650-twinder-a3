package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	once sync.Once
	zlog zerolog.Logger // Structured logger
)

// Lazy initialize loggers
func setupLogger() {
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

// Get the logger singleton
func GetLogger() zerolog.Logger {
	once.Do(setupLogger)
	return zlog
}
