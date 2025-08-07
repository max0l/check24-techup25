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

type OfferWithHotel struct {
	Offer *ent.Offer
	Hotel *ent.Hotel
}

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

func QueryOffers(request data.OfferQuerry, airports []string, client *ent.Client, ctx context.Context) []data.OfferResponse {
	results, err := client.Offer.Query().
		Where(offer.OutbounddepartureairportIn(airports...),
			offer.OutbounddeparturedatetimeGTE(request.EarliestDepartureDate),
			offer.InboundarrivaldatetimeLTE(request.LatestReturnDate),
			offer.CountadultsEQ(request.CountAdults),
			offer.CountchildrenEQ(request.CountChildren),
			offer.DurationInDaysEQ(request.Duration)).
		Order(offer.ByPrice()).
		All(ctx)
	//TODO: add Duration
	log.Debug().Msgf("Querying DB for offers done, got %v results", len(results))
	if err != nil {
		log.Warn().Err(err).Msg("Could not query DB")
		return []data.OfferResponse{}
	}

	uniqueOffers := make([]OfferWithHotel, 0, len(results))
	seenHotels := make(map[int]bool)

	for _, result := range results {
		offerHotel, err := result.QueryHotelid().Only(ctx)
		if err != nil {
			log.Warn().Err(err).Msgf("Could not query hotel for offer with ID %d", result.ID)
			continue
		}
		if !seenHotels[offerHotel.ID] {
			uniqueOffers = append(uniqueOffers, OfferWithHotel{
				Offer: result,
				Hotel: offerHotel,
			})
			seenHotels[offerHotel.ID] = true
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

		offerResult := uniqueOffers[i].Offer
		hotelResult := uniqueOffers[i].Hotel

		log.Debug().Msgf("Processing offer: %v", i)
		//TODO: ask if the hotelname should be queried -> yes, but just in case-> Fallback function with env

		response := data.OfferResponse{
			HotelName:                 hotelResult.Hotelname,
			HotelStars:                hotelResult.Hotelstars,
			Outbounddeparturedatetime: offerResult.Outbounddeparturedatetime,
			Inbounddeparturedatetime:  offerResult.Inbounddeparturedatetime,
			Countadults:               offerResult.Countadults,
			Countchildren:             offerResult.Countchildren,
			Price:                     offerResult.Price,
			Inbounddepartureairport:   offerResult.Inbounddepartureairport,
			Inboundarrivalairport:     offerResult.Inboundarrivalairport,
			Inboundarrivaldatetime:    offerResult.Inboundarrivaldatetime,
			Outbounddepartureairport:  offerResult.Outbounddepartureairport,
			Outboundarrivalairport:    offerResult.Outboundarrivalairport,
			Outboundarrivaldatetime:   offerResult.Outboundarrivaldatetime,
			Mealtye:                   offerResult.Mealtype,
			Ocenview:                  offerResult.Oceanview,
			Rommtype:                  offerResult.Roomtype,
		}
		responses = append(responses, response)
	}
	log.Debug().Msg("Built response for offers")
	return responses
}

func QueryHotel(hotelID int, client *ent.Client, ctx context.Context) (data.HotelResponse, error) {

	var hotelResponse data.HotelResponse

	log.Debug().Msgf("Querying DB for hotel with ID: %d", hotelID)
	h, err := client.Hotel.Query().
		Where(hotel.IDEQ(hotelID)).
		First(ctx)
	if err != nil {
		log.Error().Err(err).Msg("There was an error querring the DB")
		return hotelResponse, err
	}

	hotelOffers, err := client.Hotel.QueryOffers(h).All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("There was an error querring the DB")
		return hotelResponse, err
	}

	log.Info().Msgf("Found %d offers for hotel with ID: %d", len(hotelOffers), hotelID)

	hotelResponse.Hotelid = h.ID
	hotelResponse.Hotelname = h.Hotelname
	hotelResponse.Hotelstars = h.Hotelstars
	hotelResponse.Offers = make([]data.Offer, 0, len(hotelOffers))

	for i, singleOffer := range hotelOffers {
		o := data.Offer{
			Offerid:                   singleOffer.Offerid,
			Hotelid:                   singleOffer.ID,
			Outbounddeparturedatetime: singleOffer.Outbounddeparturedatetime,
			Inbounddeparturedatetime:  singleOffer.Inbounddeparturedatetime,
			Countadults:               singleOffer.Countadults,
			Countchildren:             singleOffer.Countchildren,
			Price:                     singleOffer.Price,
			Inbounddepartureairport:   singleOffer.Inbounddepartureairport,
			Inboundarrivalairport:     singleOffer.Inboundarrivalairport,
			Inboundarrivaldatetime:    singleOffer.Inboundarrivaldatetime,
			Outbounddepartureairport:  singleOffer.Outbounddepartureairport,
			Outboundarrivalairport:    singleOffer.Outboundarrivalairport,
			Outboundarrivaldatetime:   singleOffer.Outboundarrivaldatetime,
			Mealtype:                  singleOffer.Mealtype,
			Oceanview:                 singleOffer.Oceanview,
			Roomtype:                  singleOffer.Roomtype,
		}
		log.Debug().Msgf("Processing offer %d for hotel %d: %+v", i+1, hotelID, o)
		hotelResponse.Offers = append(hotelResponse.Offers, o)
	}

	return hotelResponse, nil
}
