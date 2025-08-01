package main

import (
	"check24-techup25/data"
	"check24-techup25/db"
	"check24-techup25/db/migrate"
	"check24-techup25/logging"
	"context"
	"flag"
	"github.com/rs/zerolog/log"
	"sync"
)

func main() {

	logging.SetLogger()

	hotelFilePath := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	offersFilePath := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	dbDSN := flag.String("dsn", "host=localhost port=5432 user=user dbname=check24 password=password sslmode=disable",
		"Infos to connect to the DB")
	flag.Parse()

	client, err := db.NewClient(*dbDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to DB")
	}

	defer func() {
		if err := client.Close(); err != nil {
			log.Warn().Err(err).Msg("Could not close Connection to DB client")
		}
	}()

	if err = db.Migrate(*client); err != nil {
		log.Error().Err(err).Msg("Could Not migrate current schena")
	}

	ctx := context.Background()

	hotelsResultsChannel := make(chan *data.Hotel, data.BatchSize)
	offerResultsChannel := make(chan *data.Offer, data.BatchSize)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Debug().Msg("Printed all Hotels")
		migrate.MigrateHotels(ctx, client, hotelsResultsChannel)
	}()
	data.ReadHotelsCSVFile(*hotelFilePath, hotelsResultsChannel)

	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		migrate.MigrateOffers(ctx, client, offerResultsChannel)
	}()
	data.ReadOffersCSVFile(*offersFilePath, offerResultsChannel)

	wg.Wait()

	log.Info().Msg("Done migrating")
}
