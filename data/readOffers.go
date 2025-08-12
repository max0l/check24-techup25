package data

import (
	"math"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func ReadOffersCSVFile(offersFilePath string, offerResultsChannel chan<- *Offer) {
	offersFile, err := openFile(offersFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening offers file")
		return
	}

	readCSVFile(offersFile, offerBatchWorker, ',', offerResultsChannel)

	closeFile(offersFile)
}

func offerBatchWorker(workerChan <-chan []dataRecord, offerResultsChannel chan<- *Offer, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Info().Msg("Stopped Offer Batch Worker goroutine")
	log.Info().Msg("Started Offer Batch Worker goroutine")

	for records := range workerChan {
		processOfferBatch(records, offerResultsChannel)
	}
}

func processOfferBatch(records []dataRecord, offerResultsChannel chan<- *Offer) {
	for _, record := range records {
		func(record dataRecord) {
			defer func() {
				if r := recover(); r != nil {
					log.Error().Msgf("Recovered from panic for record %v: %v", record.id, r)
				}
			}()
			offer := Offer{
				Offerid:                   record.id,
				Hotelid:                   parseInt(record.data[0]),
				Outbounddeparturedatetime: parseTime(record.data[1]),
				Inbounddeparturedatetime:  parseTime(record.data[2]),
				Countadults:               parseInt(record.data[3]),
				Countchildren:             parseInt(record.data[4]),
				Price:                     parseFloat(record.data[5]),
				Inbounddepartureairport:   record.data[6],
				Inboundarrivalairport:     record.data[7],
				Inboundarrivaldatetime:    parseTime(record.data[8]),
				Outbounddepartureairport:  record.data[9],
				Outboundarrivalairport:    record.data[10],
				Outboundarrivaldatetime:   parseTime(record.data[11]),
				Mealtype:                  record.data[12],
				Oceanview:                 record.data[13] == "true",
				Roomtype:                  record.data[14],
			}
			offer.DurationInDays = calculateDurationInDays(offer.Outbounddeparturedatetime,
				offer.Inbounddeparturedatetime)
			offerResultsChannel <- &offer
		}(record)
	}
}

func PrintOffersResults(channel <-chan *Offer) {
	for offer := range channel {
		log.Debug().Msgf("Offer: %v", offer.Hotelid)
	}
}

func calculateDurationInDays(start, end time.Time) int {
	duration := end.Sub(start)
	days := int(math.Ceil(duration.Hours() / 24))
	log.Trace().Msgf("Calculated duration: %d days from %s to %s", days, start.Format(time.RFC3339), end.Format(time.RFC3339))
	return days
}
