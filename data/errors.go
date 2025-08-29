package data

import (
	"errors"
)

var ErrNoOffersFound = errors.New("no offers found")
var ErrSiteIsOutOfRange = errors.New("requested site is out of range")
var ErrNoAirportProvided = errors.New("no Airport(s) provided")
var ErrInvalidAirport = errors.New("invalid Airport(s)")
var ErrHotelIsNotInCSV = errors.New("hotel ID is not in CSV")
