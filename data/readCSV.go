package data

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type dataRecord struct {
	id   int
	data []string
}

const (
	maxConcurrency = 10
	BatchSize      = 10000
)

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

func readCSVFile[T any](csvFile *os.File, batchFunc func(<-chan []dataRecord, chan<- *T, *sync.WaitGroup), csvComma rune, resultsChannel chan<- *T) {
	defer close(resultsChannel)
	defer log.Info().Msg("CSV file reading completed.")
	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = csvComma

	readHeaders(csvReader)

	processorChannel := make(chan []dataRecord, maxConcurrency)
	var records []dataRecord
	log.Info().Msg("Starting to read CSV records...")
	var wg sync.WaitGroup

	// starting 10 workers
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go batchFunc(processorChannel, resultsChannel, &wg)
	}
	var i int
	for {

		line, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Error().Err(err).Msg("Error reading CSV record")
			continue
		}

		//log.Debug().Msgf("Read record: %v", line)

		record := dataRecord{
			i,
			line,
		}

		i++

		records = append(records, record)

		if len(records) >= BatchSize {
			batchCopy := make([]dataRecord, len(records))
			copy(batchCopy, records)
			processorChannel <- batchCopy
			records = records[:0]
		}
	}

	if len(records) > 0 {
		processorChannel <- records
	}

	log.Debug().Msg("Waiting for all Goriutines to finish")
	close(processorChannel)
	wg.Wait()
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
