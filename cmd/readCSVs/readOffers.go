package main

import (
	"check24-techup25/data"
	"fmt"
	"github.com/rs/zerolog/log"
)

func ReadOffersCSVFile(offersFilePath string, offerResultsChannel chan<- *data.Offer) {
	offersFile, err := openFile(offersFilePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening offers file")
		return
	}

	readCSVFile(offersFile, processOfferBatch, ',', offerResultsChannel)

	closeFile(offersFile)
}

func processOfferBatch(records [][]string, offerResultsChannel chan<- *data.Offer) {
	for _, record := range records {

		offer := data.Offer{
			Hotelid:                   parseInt(record[0]),
			Outbounddeparturedatetime: parseTime(record[1]),
			Inbounddeparturedatetime:  parseTime(record[2]),
			Countadults:               parseInt(record[3]),
			Countchildren:             parseInt(record[4]),
			Price:                     parseFloat(record[5]),
			Inbounddepartureairport:   record[6],
			Inboundarrivalairport:     record[7],
			Inboundarrivaldatetime:    parseTime(record[8]),
			Outbounddepartureairport:  record[9],
			Outboundarrivalairport:    record[10],
			Outboundarrivaldatetime:   parseTime(record[11]),
			Mealtype:                  record[12],
			Oceanview:                 record[13] == "true",
			Roomtype:                  record[14],
		}
		offerResultsChannel <- &offer
	}
}

func consumeOffersResults(channel <-chan *data.Offer) {
	for offer := range channel {
		fmt.Print(offer.Hotelid)
	}
}
