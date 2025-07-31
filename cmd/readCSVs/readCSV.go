package main

import (
	"check24-techup25/data"
	"check24-techup25/logging"
	"encoding/csv"
	"flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strconv"
	"time"
)

func main() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	hotelFilePath := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	debug := flag.Bool("debug", false, "enables debug logging")
	offersFilePath := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	logging.SetLogger(*debug)

	hotelFile, err := openFile(*hotelFilePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening hotelFilePath file")
		return
	}

	var hotels []data.Hotel

	readCSVFile(hotelFile, &hotels, processHotelBatch)

	closeFile(hotelFile)

	var offers []data.Offer

	offersFile, err := openFile(*offersFilePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening offers file")
		return
	}

	readCSVFile(offersFile, &offers, processOfferBatch)

	closeFile(offersFile)

	for _, hotel := range hotels {
		log.Debug().Msgf("Hotel ID: %d, Name: %s, Stars: %.1f\n", hotel.Hotelid, hotel.Hotelname, hotel.Hotelstars)
	}
}

func openFile(filePath string) (*os.File, error) {
	csvFile, err := os.Open(filePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening hotelFilePath file")
		return nil, err
	}

	return csvFile, nil
}

func closeFile(csvFile *os.File) {
	defer func(csvFile *os.File) {
		err := csvFile.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error closing hotelFilePath file")
		} else {
			log.Debug().Msg("Hotels file closed successfully.")
		}
	}(csvFile)
}

func readCSVFile[T any](csvFile *os.File, dest *[]T, batchFunc func([][]string, *[]T)) {
	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = ';'

	readHeaders(csvReader)

	const batchSize = 100
	var records [][]string

	log.Info().Msg("Starting to read CSV records...")

	for i := 0; i < 1000; i++ {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Error().Err(err).Msg("Error reading CSV record")
			continue
		}

		log.Info().Msgf("Read record: %v", record)

		records = append(records, record)

		if len(records) >= batchSize {
			batchFunc(records, dest)
			records = records[:0]
		}
	}
}

func readHeaders(csvReader *csv.Reader) {
	headers, err := csvReader.Read()
	if err != nil {
		log.Panic().Err(err).Msg("Error reading header from hotelFilePath file")
	}

	log.Debug().Msgf("CSV Headers: %v", headers)
}

func parseInt(str string) int {
	value, err := strconv.Atoi(str)
	if err != nil {
		log.Error().Err(err).Msgf("Error converting string to int: %s", str)
		return 0
	}
	return value
}

func parseFloat(str string) float32 {
	value, err := strconv.ParseFloat(str, 32)
	if err != nil {
		log.Error().Err(err).Msgf("Error converting string to float: %s", str)
		return 0.0
	}
	return float32(value)
}

func parseTime(dateStr string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing date string: %s", dateStr)
		return time.Time{}
	}
	return parsedTime
}
