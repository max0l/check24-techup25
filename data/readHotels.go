package data

import (
	"sync"

	"github.com/rs/zerolog/log"
)

func ReadHotelsCSVFile(hotelFilePath string, hotelResultsChannel chan<- *Hotel) {

	hotelFile, err := openFile(hotelFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening hotelFilePath file")
		return
	}

	readCSVFile(hotelFile, hotelBatchWorker, ';', hotelResultsChannel)

	closeFile(hotelFile)
}

func hotelBatchWorker(workerChan <-chan []dataRecord, hotelResultsChannel chan<- *Hotel, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Info().Msg("Stopped Hotel Batch Worker goroutine")
	log.Info().Msg("Started Hotel Batch Worker goroutine")

	for records := range workerChan {
		processHotelBatch(records, hotelResultsChannel)
	}
}

func processHotelBatch(records []dataRecord, hotelResultsChannel chan<- *Hotel) {
	for _, record := range records {
		func(record dataRecord) {
			defer func() {
				if r := recover(); r != nil {
					log.Error().Msgf("Recovered from panic for record %v: %v", record.id, r)
				}
			}()
			hotel := Hotel{
				Hotelid:    parseInt(record.data[0]),
				Hotelname:  record.data[1],
				Hotelstars: parseFloat(record.data[2]),
			}
			hotelResultsChannel <- &hotel
		}(record)
	}
}

func PrintHotelsResults(channel <-chan *Hotel) {
	for hotel := range channel {
		log.Debug().Msgf("%v", hotel.Hotelid)
	}
}
