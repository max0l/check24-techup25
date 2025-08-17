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

type RequestQuerryParams struct {
	DepartureAirports     []string  `form:"departureAirports"`
	EarliestDepartureDate time.Time `form:"earliestDepartureDate"`
	LatestReturnDate      time.Time `form:"latestReturnDate"`
	CountAdults           int       `form:"countAdults"`
	CountChildren         int       `form:"countChildren"`
	Duration              int       `form:"duration"`
	Site                  int       `form:"site"`
}

type OfferWithHotel struct {
	HotelName                 string    `json:"hotelname" example:"Hotel Sunshine"`
	HotelStars                float32   `json:"hotelstars" example:"4.5"`
	HotelId                   int       `json:"hotelid" example:"123"`
	DurationInDays            int       `json:"durationindays" example:"7"`
	Outbounddeparturedatetime time.Time `json:"outbounddeparturedatetime" example:"2025-01-01T00:00:00+02:00"`
	Inbounddeparturedatetime  time.Time `json:"inbounddeparturedatetime" example:"2025-12-31T00:00:00+02:00"`
	Countadults               int       `json:"countadults" example:"2"`
	Countchildren             int       `json:"countchildren" example:"1"`
	Price                     float32   `json:"price" example:"299.99"`
	Inbounddepartureairport   string    `json:"inbounddepartureairport" example:"BER"`
	Inboundarrivalairport     string    `json:"inboundarrivalairport" example:"PMI"`
	Inboundarrivaldatetime    time.Time `json:"inboundarrivaldatetime" example:"2025-12-31T00:00:00+02:00"`
	Outbounddepartureairport  string    `json:"outbounddepartureairport" example:"PMI"`
	Outboundarrivalairport    string    `json:"outboundarrivalairport" example:"BER"`
	Outboundarrivaldatetime   time.Time `json:"outboundarrivaldatetime" example:"2025-01-01T00:00:00+02:00"`
	Mealtye                   string    `json:"mealtye" example:"allinclusive"`
	Ocenview                  bool      `json:"ocenview" example:"true"`
	Rommtype                  string    `json:"rommtype" example:"double"`
}

type OfferResponse struct {
	Offers      []OfferWithHotel `json:"offers"`
	CountOffers int              `json:"countOffers" example:"5"`
}
