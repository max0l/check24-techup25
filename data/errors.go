package data

import "errors"

var ErrNoOffersFound = errors.New("no offers found")
var ErrSiteIsOutOfRange = errors.New("requested site is out of range")
