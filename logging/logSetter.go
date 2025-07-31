package logging

import (
	"github.com/rs/zerolog"
)

func SetLogger(debug bool) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
}
