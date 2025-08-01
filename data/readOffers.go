package data

import (
	"github.com/rs/zerolog/log"
	"sync"
)

func ReadOffersCSVFile(offersFilePath string, offerResultsChannel chan<- *Offer) {
	offersFile, err := openFile(offersFilePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening offers file")
		return
	}

	readCSVFile(offersFile, offerBatchWorker, ',', offerResultsChannel)

	closeFile(offersFile)
}

func offerBatchWorker(workerChan <-chan [][]string, offerResultsChannel chan<- *Offer, wg *sync.WaitGroup) {
	defer wg.Done()

	for records := range workerChan {
		processOfferBatch(records, offerResultsChannel)
	}
}

func processOfferBatch(records [][]string, offerResultsChannel chan<- *Offer) {
	for _, record := range records {

		offer := Offer{
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

func PrintOffersResults(channel <-chan *Offer) {
	for offer := range channel {
		log.Debug().Msgf("Offer: %v", offer.Hotelid)
	}
}
