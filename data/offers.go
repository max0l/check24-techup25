package data

import "time"

type Offer struct {
	Offerid                   int       `json:"offerid"`
	Hotelid                   int       `json:"hotelid"`
	Outbounddeparturedatetime time.Time `json:"outbounddeparturedatetime"`
	Inbounddeparturedatetime  time.Time `json:"inbounddeparturedatetime"`
	Countadults               int       `json:"countadults"`
	Countchildren             int       `json:"countchildren"`
	Price                     float32   `json:"price"`
	Inbounddepartureairport   string    `json:"inbounddepartureairport"`
	Inboundarrivalairport     string    `json:"inboundarrivalairport"`
	Inboundarrivaldatetime    time.Time `json:"inboundarrivaldatetime"`
	Outbounddepartureairport  string    `json:"outbounddepartureairport"`
	Outboundarrivalairport    string    `json:"outboundarrivalairport"`
	Outboundarrivaldatetime   time.Time `json:"outboundarrivaldatetime"`
	Mealtype                  string    `json:"mealtype"`
	Oceanview                 bool      `json:"oceanview"`
	Roomtype                  string    `json:"roomtype"`
	DurationInDays            int       `json:"durationindays"`
}

type RequestQuerryParams struct {
	DepartureAirports     []string  `form:"departureAirports" url:"departureAirports,omitempty" del:","`
	EarliestDepartureDate time.Time `form:"earliestDepartureDate" url:"earliestDepartureDate,omitempty"`
	LatestReturnDate      time.Time `form:"latestReturnDate" url:"latestReturnDate,omitempty"`
	CountAdults           int       `form:"countAdults" url:"countAdults,omitempty"`
	CountChildren         int       `form:"countChildren" url:"countChildren,omitempty"`
	Duration              int       `form:"duration" url:"duration,omitempty"`
	Site                  int       `form:"site" url:"site,omitempty"`
}

type BestOfferPerHotel struct {
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
	OfferAmount               int       `json:"offeramount" example:"1"`
}

type BestOfferPerHotelResponse struct {
	Offers      []BestOfferPerHotel `json:"offers"`
	CountOffers int                 `json:"countOffers" example:"5"`
}
