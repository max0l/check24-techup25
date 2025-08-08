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
	DurationInDays            int
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
	HotelName                 string    `json:"HotelName" example:"Hotel Sunshine"`
	HotelStars                float32   `json:"HotelStars" example:"4.5"`
	Outbounddeparturedatetime time.Time `json:"Outbounddeparturedatetime" example:"2025-01-01T00:00:00+02:00"`
	Inbounddeparturedatetime  time.Time `json:"Inbounddeparturedatetime" example:"2025-12-31T00:00:00+02:00"`
	Countadults               int       `json:"Countadults" example:"2"`
	Countchildren             int       `json:"Countchildren" example:"1"`
	Price                     float32   `json:"Price" example:"299.99"`
	Inbounddepartureairport   string    `json:"Inbounddepartureairport" example:"BER"`
	Inboundarrivalairport     string    `json:"Inboundarrivalairport" example:"PMI"`
	Inboundarrivaldatetime    time.Time `json:"Inboundarrivaldatetime" example:"2025-12-31T00:00:00+02:00"`
	Outbounddepartureairport  string    `json:"Outbounddepartureairport" example:"PMI"`
	Outboundarrivalairport    string    `json:"Outboundarrivalairport" example:"BER"`
	Outboundarrivaldatetime   time.Time `json:"Outboundarrivaldatetime" example:"2025-01-01T00:00:00+02:00"`
	Mealtye                   string    `json:"Mealtye" example:"allinclusive"`
	Ocenview                  bool      `json:"Ocenview" example:"true"`
	Rommtype                  string    `json:"Rommtype" example:"double"`
}
