package data

import "time"

type Offer struct {
	Offerid                   int
	Hotelid                   int
	Outbounddeparturedatetime time.Time
	Inbounddeparturedatetime  time.Time
	Countadults               int
	Countchildren             int
	Price                     float32
	Inbounddepartureairport   string
	Inboundarrivalairport     string
	Inboundarrivaldatetime    time.Time
	Outbounddepartureairport  string
	Outboundarrivalairport    string
	Outboundarrivaldatetime   time.Time
	Mealtype                  string
	Oceanview                 bool
	Roomtype                  string
}

type OfferQuerry struct {
	DepartureAirports     []string  `binding:"required"`
	EarliestDepartureDate time.Time `binding:"required"`
	LatestReturnDate      time.Time `binding:"required"`
	CountAdults           int       `binding:"required"`
	CountChildren         int
	Duration              int `binding:"required"`
	Site                  int `binding:"required"`
}

type OfferResponse struct {
	HotelName                 string
	HotelStars                float32
	Outbounddeparturedatetime time.Time
	Inbounddeparturedatetime  time.Time
	Countadults               int
	Countchildren             int
	Price                     float32
	Inbounddepartureairport   string
	Inboundarrivalairport     string
	Inboundarrivaldatetime    time.Time
	Outbounddepartureairport  string
	Outboundarrivalairport    string
	Outboundarrivaldatetime   time.Time
	Mealtye                   string
	Ocenview                  bool
	Rommtype                  string
}
