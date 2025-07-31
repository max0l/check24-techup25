package main

import (
	"check24-techup25/data"
	"github.com/rs/zerolog/log"
)

func ReadHotelsCSVFile(hotelFilePath string, hotelResultsChannel chan<- *data.Hotel) {

	hotelFile, err := openFile(hotelFilePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening hotelFilePath file")
		return
	}

	readCSVFile(hotelFile, processHotelBatch, ';', hotelResultsChannel)

	closeFile(hotelFile)

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
