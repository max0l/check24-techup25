package main

import (
	"check24-techup25/data"
	"check24-techup25/logging"
	"flag"
	"sync"

	"github.com/rs/zerolog/log"
)

func main() {

	hotelFilePath := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	offersFilePath := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	logging.SetLogger()

	hotelsResultsChannel := make(chan *data.Hotel, data.BatchSize)
	offerResultsChannel := make(chan *data.Offer, data.BatchSize)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Debug().Msg("Printed all Hotels")
		data.PrintHotelsResults(hotelsResultsChannel)
	}()
	data.ReadHotelsCSVFile(*hotelFilePath, hotelsResultsChannel)

	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		data.PrintOffersResults(offerResultsChannel)
	}()
	data.ReadOffersCSVFile(*offersFilePath, offerResultsChannel)

	wg.Wait()
}
