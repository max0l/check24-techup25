package main

import (
	"check24-techup25/data"
	"check24-techup25/logging"
	"encoding/csv"
	"flag"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strconv"
	"time"
)

func main() {

	hotelFilePath := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	offersFilePath := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	logging.SetLogger()

	hotelsResultsChannel := make(chan<- *data.Hotel, 10000)

	ReadHotelsCSVFile(*hotelFilePath, hotelsResultsChannel)

	offerResultsChannel := make(chan<- *data.Offer, 10000)

	ReadOffersCSVFile(*offersFilePath, offerResultsChannel)

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

func readCSVFile[T any](csvFile *os.File, batchFunc func([][]string, chan<- *T), csvComma rune, resultsChannel chan<- *T) {
	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = csvComma

	readHeaders(csvReader)

	const batchSize = 10000
	var records [][]string

	log.Info().Msg("Starting to read CSV records...")

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Error().Err(err).Msg("Error reading CSV record")
			continue
		}

		log.Debug().Msgf("Read record: %v", record)

		records = append(records, record)

		if len(records) >= batchSize {
			batchFunc(records, resultsChannel)
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
