package logger

import (
	"os"
	"strings"

	recipe "github.com/cunkz/goyummy/bin/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger(cfg *recipe.AppRecipe) {
	// Set log output (stdout or file)
	var output *os.File
	if cfg.Logging.Output == "stdout" {
		output = os.Stdout
	} else {
		// file log (plain text)
		f, err := os.OpenFile(cfg.Logging.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal().Err(err).Msg("Cannot open log file")
		}
		output = f
	}

	// set level
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}
