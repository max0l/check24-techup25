package main

import (
	"check24-techup25/data"
	"check24-techup25/logging"
	"flag"
	"sync"
)

func main() {

	hotelFilePath := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	offersFilePath := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	logging.SetLogger()

	hotelsResultsChannel := make(chan *data.Hotel, 10000)
	offerResultsChannel := make(chan *data.Offer, 10000)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ReadHotelsCSVFile(*hotelFilePath, hotelsResultsChannel)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		printHotelsResults(hotelsResultsChannel)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ReadOffersCSVFile(*offersFilePath, offerResultsChannel)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		printOffersResults(offerResultsChannel)
	}()

	wg.Wait()
}
