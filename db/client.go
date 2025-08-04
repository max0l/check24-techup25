package db

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"check24-techup25/db/ent/hotel"
	"check24-techup25/db/ent/offer"
	"context"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// NewClient initializes and returns a new Ent client for a PostgreSQL database.
// Example DSN: "host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable"
func NewClient(dsn string) (*ent.Client, error) {
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func Migrate(client ent.Client) error {
	if err := client.Schema.Create(context.Background()); err != nil {
		return err
	}
	return nil
}

func QueryOffers(request data.OfferQuerry, client *ent.Client, ctx context.Context) []data.OfferResponse {
	results, err := client.Offer.Query().
		Where(offer.OutbounddepartureairportIn(request.DepartureAirports...),
			offer.OutbounddeparturedatetimeGT(request.EarliestDepartureDate),
			offer.Inboundarrivaldatetime(request.LatestReturnDate),
			offer.CountadultsEQ(request.CountAdults),
			offer.Countchildren(request.CountChildren)).
		Order(offer.ByPrice()).
		All(ctx)
	//TODO: add Duration
	if err != nil {
		log.Warn().Err(err).Msg("Could not query DB")
		return []data.OfferResponse{}
	}

	var responses []data.OfferResponse
	for _, result := range results {

		queriedHotel, err := client.Hotel.Query().
			Where(hotel.HotelidEQ(result.Hotelid)).
			First(ctx)
		if err != nil {
			continue
		}

		response := data.OfferResponse{
			HotelName:                 queriedHotel.Hotelname,
			Outbounddeparturedatetime: result.Outbounddeparturedatetime,
			Inbounddeparturedatetime:  result.Inbounddeparturedatetime,
			Countadults:               result.Countadults,
			Countchildren:             result.Countchildren,
			Price:                     result.Price,
			Inbounddepartureairport:   result.Inbounddepartureairport,
			Inboundarrivalairport:     result.Inboundarrivalairport,
			Inboundarrivaldatetime:    result.Inboundarrivaldatetime,
			Outbounddepartureairport:  result.Outbounddepartureairport,
			Outboundarrivalairport:    result.Outboundarrivalairport,
			Outboundarrivaldatetime:   result.Outboundarrivaldatetime,
			Mealtye:                   result.Mealtype,
			Ocenview:                  result.Oceanview,
			Rommtype:                  result.Roomtype,
		}
		responses = append(responses, response)
	}

	return responses
}
