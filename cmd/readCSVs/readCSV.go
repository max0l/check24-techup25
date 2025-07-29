package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {

	hotels := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	//offers := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	hotelFile, err := os.Open(*hotels)
	if err != nil {
		panic("Error opening hotels file: " + err.Error())
	}
	defer func(hotelFile *os.File) {
		err := hotelFile.Close()
		if err != nil {
			fmt.Println("Error closing hotels file:", err)
		} else {
			fmt.Println("Hotels file closed successfully.")
		}
	}(hotelFile)

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
				panic("Error reading hotels file: " + err.Error())
			}

		}
		fmt.Print(string(buf))
	}
}
