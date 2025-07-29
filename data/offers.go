package data

import "time"

type Offer struct {
	hotelid                   int
	outbounddeparturedatetime time.Time
	inbounddeparturedatetime  time.Time
	countadults               int
	countchildren             int
	price                     float32
	inbounddepartureairport   string
	inboundarrivalairport     string
	inboundarrivaldatetime    time.Time
	outbounddepartureairport  string
	outboundarrivalairport    string
	outboundarrivaldatetime   time.Time
	mealtype                  string
	oceanview                 bool
	roomtype                  string
}
