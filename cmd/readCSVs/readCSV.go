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

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	hotelFilePath := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	//offers := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	hotelFile, err := os.Open(*hotelFilePath)
	if err != nil {
		panic("Error opening hotelFilePath file: " + err.Error())
	}
	defer func(hotelFile *os.File) {
		err := hotelFile.Close()
		if err != nil {
			fmt.Println("Error closing hotelFilePath file:", err)
		} else {
			fmt.Println("Hotels file closed successfully.")
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
				panic("Error reading hotelFilePath file: " + err.Error())
			}

		}

		if err := gocsv.UnmarshalBytes(buf, &hotels); err != nil {
			panic("Error unmarshalling hotelFilePath file: " + err.Error())
		}
	}
	for _, hotel := range hotels {
		log.Debug().Msgf("Hotel ID: %d, Name: %s, Stars: %.1f\n", hotel.Hotelid, hotel.Hotelname, hotel.Hotelstars)
	}
}
