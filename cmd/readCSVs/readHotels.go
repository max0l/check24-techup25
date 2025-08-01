package main

import (
	"check24-techup25/data"
	"github.com/rs/zerolog/log"
	"sync"
)

func ReadHotelsCSVFile(hotelFilePath string, hotelResultsChannel chan<- *data.Hotel) {

	hotelFile, err := openFile(hotelFilePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening hotelFilePath file")
		return
	}

	readCSVFile(hotelFile, hotelBatchWorker, ';', hotelResultsChannel)

	closeFile(hotelFile)
}

func hotelBatchWorker(workerChan <-chan [][]string, hotelResultsChannel chan<- *data.Hotel, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Info().Msg("Stopped Hotel Batch Worker goroutine")
	log.Info().Msg("Started Hotel Batch Worker goroutine")

	for records := range workerChan {
		processHotelBatch(records, hotelResultsChannel)
	}
}

func processHotelBatch(records [][]string, hotelResultsChannel chan<- *data.Hotel) {
	for _, record := range records {

		hotel := data.Hotel{
			Hotelid:    parseInt(record[0]),
			Hotelname:  record[1],
			Hotelstars: parseFloat(record[2]),
		}
		hotelResultsChannel <- &hotel
	}
}

func printHotelsResults(channel <-chan *data.Hotel) {
	for hotel := range channel {
		log.Debug().Msgf("%v", hotel.Hotelid)
	}
}
