package readCSVs

import "flag"

func main() {

	hotels := flag.String("h", "./data/hotels.csv", "Path to the hotels CSV file")
	offers := flag.String("o", "../offers/offer.csv", "Path to the offers CSV file")
	flag.Parse()

	readHotels(*hotels)
	readOffers(*offers)
}
