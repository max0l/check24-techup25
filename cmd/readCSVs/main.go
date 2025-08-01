package main

import (
	"check24-techup25/data"
	"check24-techup25/logging"
	"flag"
)

func main() {

	hotelFilePath := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	offersFilePath := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	logging.SetLogger()

	hotelsResultsChannel := make(chan *data.Hotel, batchSize)
	offerResultsChannel := make(chan *data.Offer, batchSize)
	startReader(hotelFilePath, hotelsResultsChannel, offersFilePath, offerResultsChannel)
}
