package main

import (
	"check24-techup25/data"
)

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
