package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetLogger(debug bool) {
	if debug {
		log.Info().Msg("Debug logging enabled")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		log.Info().Msg("Debug logging disabled")
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
}
