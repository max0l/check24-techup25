package main

import (
	"check24-techup25/data"
	"check24-techup25/db"
	"check24-techup25/db/ent"
	"check24-techup25/db/migrate"
	"check24-techup25/logging"
	"context"
	"flag"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
)

// main function initializes the database connection and starts the migration process.
// It reads the CSV file and upsert the database with the hotel and offer data.
func main() {

	logging.SetLogger()

	hotelFilePath := flag.String("hotel", "./data/hotels.csv", "Path to the hotels CSV file")
	offersFilePath := flag.String("offer", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	shouldSeed := os.Getenv("SEED_DATA")
	if shouldSeed == "false" {
		log.Info().Msg("SEED_DATA is set to false, skipping data migration")
		return
	}

	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN == "" {
		dbDSN = "host=localhost port=5432 user=user dbname=check24 password=password sslmode=disable"
	}

	client := parseDbDSN(dbDSN)

	defer func() {
		if err := client.Close(); err != nil {
			log.Warn().Err(err).Msg("Could not close Connection to DB client")
		}
	}()

	migrateSchema(client)

	ctx := context.Background()

	pingDB(client, ctx)

	hotelsResultsChannel := make(chan *data.Hotel, data.BatchSize)
	offerResultsChannel := make(chan *data.Offer, data.BatchSize)

	log.Info().Msg("Start Reading CSV's...")

	var wg sync.WaitGroup

	startMigrating(&wg, ctx, client, hotelsResultsChannel, hotelFilePath, offerResultsChannel, offersFilePath)

	wg.Wait()

	log.Info().Msg("Done")
}

func parseDbDSN(dbDSN string) *ent.Client {
	log.Info().Msg("Connecting to DB...")
	client, err := db.NewClient(dbDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to DB")
	} else {
		log.Info().Msg("Connected to DB")
	}
	return client
}

func migrateSchema(client *ent.Client) {
	log.Info().Msg("Running Migration...")
	if err := db.Migrate(*client); err != nil {
		log.Error().Err(err).Msg("Could Not migrate current schema")
	} else {
		log.Info().Msg("Migration Done")
	}
}

func pingDB(client *ent.Client, ctx context.Context) {
	if err := client.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Could not ping DB")
	} else {
		log.Info().Msg("Ping to DB successful")
	}
}

func startMigrating(wg *sync.WaitGroup, ctx context.Context, client *ent.Client, hotelsResultsChannel chan *data.Hotel,
	hotelFilePath *string, offerResultsChannel chan *data.Offer, offersFilePath *string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Debug().Msg("Processed all Hotels")
		migrate.MigrateHotels(ctx, client, hotelsResultsChannel)
	}()
	data.ReadHotelsCSVFile(*hotelFilePath, hotelsResultsChannel)

	wg.Wait()

	knownHotels := migrate.NewKnownHotels()

	wg.Add(1)
	go func() {
		defer wg.Done()
		migrate.MigrateOffers(ctx, client, knownHotels, offerResultsChannel)
	}()
	data.ReadOffersCSVFile(*offersFilePath, offerResultsChannel)
	return
}
