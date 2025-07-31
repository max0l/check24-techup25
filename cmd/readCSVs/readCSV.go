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
	//offers := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	logging.SetLogger(*debug)

	hotelFile, err := openFile(*hotelFilePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening hotelFilePath file")
		return
	}

	var hotels []data.Hotel

	csvReader := csv.NewReader(hotelFile)
	csvReader.Comma = ';'

	// Read header first
	headers, err := csvReader.Read()
	if err != nil {
		log.Panic().Err(err).Msg("Error reading header from hotelFilePath file")
	}

	log.Debug().Msgf("CSV Headers: %v", headers)

	const batchSize = 10
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

		log.Info().Msgf("Read record: %v", record)

		records = append(records, record)

		if len(records) >= batchSize {
			processHotelBatch(records, &hotels)
			records = records[:0]
		}
	}

	for _, hotel := range hotels {
		log.Debug().Msgf("Hotel ID: %d, Name: %s, Stars: %.1f\n", hotel.Hotelid, hotel.Hotelname, hotel.Hotelstars)
	}
}

func openFile(filePath string) (*os.File, error) {
	hotelFile, err := os.Open(filePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening hotelFilePath file")
		return nil, err
	}
	defer func(hotelFile *os.File) {
		err := hotelFile.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error closing hotelFilePath file")
		} else {
			log.Debug().Msg("Hotels file closed successfully.")
		}
	}(hotelFile)

	return hotelFile, nil
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
