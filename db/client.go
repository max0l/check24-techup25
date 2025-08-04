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
			offer.OutbounddeparturedatetimeGTE(request.EarliestDepartureDate),
			offer.InboundarrivaldatetimeLTE(request.LatestReturnDate),
			offer.CountadultsEQ(request.CountAdults),
			offer.CountchildrenEQ(request.CountChildren)).
		Order(offer.ByPrice()).
		Limit(50).
		All(ctx)
	//TODO: add Duration
	log.Debug().Msgf("Querying DB for offers done, got %v results", len(results))
	if err != nil {
		log.Warn().Err(err).Msg("Could not query DB")
		return []data.OfferResponse{}
	}

	var responses []data.OfferResponse
	for i, result := range results {
		log.Debug().Msgf("Processing offer: %v", i)
		queriedHotel, err := client.Hotel.Query().
			Where(hotel.HotelidEQ(result.Hotelid)).
			First(ctx)
		if err != nil {
			continue
		}

		response := data.OfferResponse{
			HotelName:                 queriedHotel.Hotelname,
			HotelStars:                queriedHotel.Hotelstars,
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
	log.Debug().Msg("Built response for offers")
	return responses
}
