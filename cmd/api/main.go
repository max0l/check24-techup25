package main

import (
	"check24-techup25/api"
	"check24-techup25/db"
	"check24-techup25/logging"
	"context"
	"flag"
	"github.com/rs/zerolog/log"
)

func main() {
	logging.SetLogger()
	port := flag.Int("port", 8080, "Set the port of the API")
	dbDSN := flag.String("dsn", "host=localhost port=5432 user=user dbname=check24 password=password sslmode=disable",
		"Infos to connect to the DB")
	flag.Parse()

	log.Info().Msg("Connecting to DB...")
	client, err := db.NewClient(*dbDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("DSN is not valid")
	} else {
		log.Info().Msg("Valid DSN")
	}

	defer func() {
		if err := client.Close(); err != nil {
			log.Warn().Err(err).Msg("Could not close Connection to DB client")
		}
	}()
	ctx := context.Background()

	if err := client.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Could not ping DB")
	} else {
		log.Info().Msg("Ping to DB successful")
	}

	api.Start(*port, client, ctx)
}
