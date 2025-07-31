package main

import (
	"check24-techup25/data"
	"github.com/rs/zerolog/log"
)

func readHotelsCSVFile(hotelFilePath string, hotels *[]data.Hotel) {

	hotelFile, err := openFile(hotelFilePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening hotelFilePath file")
		return
	}

	readCSVFile(hotelFile, hotels, processHotelBatch)

	closeFile(hotelFile)

}

func processHotelBatch(records [][]string, hotels *[]data.Hotel) {
	for _, record := range records {

		hotel := data.Hotel{
			Hotelid:    parseInt(record[0]),
			Hotelname:  record[1],
			Hotelstars: parseFloat(record[2]),
		}
		*hotels = append(*hotels, hotel)
	}
}
