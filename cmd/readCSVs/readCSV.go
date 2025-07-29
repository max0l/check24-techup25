package main

import (
	"bufio"
	"check24-techup25/data"
	"flag"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
)

func main() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	hotelFilePath := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	debug := flag.Bool("debug", false, "sets log level to debug")
	//offers := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	hotelFile, err := os.Open(*hotelFilePath)
	if err != nil {
		log.Panic().Err(err).Msg("Error opening hotelFilePath file")
	}
	defer func(hotelFile *os.File) {
		err := hotelFile.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error closing hotelFilePath file")
		} else {
			log.Debug().Msg("Hotels file closed successfully.")
		}
	}(hotelFile)

	var hotels []*data.Hotel

	r := bufio.NewReader(hotelFile)
	for {
		buf := make([]byte, 4*1024)
		n, err := r.Read(buf)
		buf = buf[:n]
		if n == 0 {
			if err != nil {
				fmt.Println(err)
				break
			}
			if err == io.EOF {
				break
			} else if err != io.EOF {
				log.Panic().Err(err).Msg("Error reading hotelFilePath file")
			}

		}

		if err := gocsv.UnmarshalBytes(buf, &hotels); err != nil {
			log.Error().Err(err).Msg("Error unmarshalling hotelFilePath file")
		}
	}
	for _, hotel := range hotels {
		log.Debug().Msgf("Hotel ID: %d, Name: %s, Stars: %.1f\n", hotel.Hotelid, hotel.Hotelname, hotel.Hotelstars)
	}
}
