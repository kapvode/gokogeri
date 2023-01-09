package gokogeri

import (
	"os"

	"github.com/rs/zerolog"
)

// DefaultLogger returns a JSON logger, writing to stdout, with a timestamp and a minimum level of info.
func DefaultLogger() zerolog.Logger {
	return zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
}
