package db

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"check24-techup25/db/ent/hotel"
	"check24-techup25/db/ent/offer"
	"context"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"strconv"
)

const responseSize = 50

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
		All(ctx)
	//TODO: add Duration
	log.Debug().Msgf("Querying DB for offers done, got %v results", len(results))
	if err != nil {
		log.Warn().Err(err).Msg("Could not query DB")
		return []data.OfferResponse{}
	}

	uniqueOffers := make([]*ent.Offer, 0, len(results))
	seenHotels := make(map[int]bool)

	for _, result := range results {
		hotelID := result.Hotelid
		if !seenHotels[hotelID] {
			uniqueOffers = append(uniqueOffers, result)
			seenHotels[hotelID] = true
		}
	}

	log.Debug().Msgf("Found %v unique offers", len(uniqueOffers))

	processedResults := 0
	var responses []data.OfferResponse

	if responseSize*(request.Site-1) >= len(uniqueOffers) {
		log.Warn().Msg("Response size limit is larger than available offers")
		return []data.OfferResponse{}
	}

	for i := responseSize * (request.Site - 1); i < len(uniqueOffers); i++ {

		if processedResults >= responseSize {
			log.Debug().Msg("Reached response size limit, stopping processing offers")
			break
		}
		processedResults++

		result := uniqueOffers[i]

		log.Debug().Msgf("Processing offer: %v", i)
		//TODO: ask if the hotelname should be queried
		queriedHotel, err := client.Hotel.Query().
			Where(hotel.HotelidEQ(result.Hotelid)).
			First(ctx)
		if err != nil {
			log.Warn().Msgf("Could not find hotel with ID %v for offer %v. "+
				"Falling back to ID's", result.Hotelid, result.Offerid)
			queriedHotel = &ent.Hotel{
				Hotelname:  strconv.Itoa(result.Hotelid),
				Hotelstars: 0,
			}
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
