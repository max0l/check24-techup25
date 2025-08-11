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

const ResponseSize = 50

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

// TODO: change edge name 'hotelid' of offer
func QueryOffers(request data.OfferQuerry, airports []string, client *ent.Client, ctx context.Context) []*ent.Offer {
	results, err := client.Debug().Offer.Query().
		Where(offer.OutbounddepartureairportIn(airports...),
			offer.OutbounddeparturedatetimeGTE(request.EarliestDepartureDate),
			offer.InboundarrivaldatetimeLTE(request.LatestReturnDate),
			offer.CountadultsEQ(request.CountAdults),
			offer.CountchildrenEQ(request.CountChildren),
			offer.DurationInDaysEQ(request.Duration)).
		Order(offer.ByPrice()).WithHotel().
		All(ctx)
	//TODO: add Duration
	log.Debug().Msgf("Querying DB for offers done, got %v results", len(results))
	if err != nil {
		log.Warn().Err(err).Msg("Could not query DB")
		return []*ent.Offer{}
	}

	uniqueOffers := make([]*ent.Offer, 0, len(results))
	seenHotels := make(map[int]bool)

	for _, result := range results {
		offerHotelId := result.Edges.Hotel.ID
		if !seenHotels[offerHotelId] {
			uniqueOffers = append(uniqueOffers, result)
			seenHotels[offerHotelId] = true
		}
	}

	log.Debug().Msgf("Found %v unique offers", len(uniqueOffers))

	if ResponseSize*(request.Site-1) >= len(uniqueOffers) {
		log.Warn().Msg("Response size limit is larger than available offers")
		return []*ent.Offer{}
	}

	log.Debug().Msg("Queried offers")
	return uniqueOffers
}

func QueryHotel(hotelID int, site int, client *ent.Client, ctx context.Context) (data.HotelResponse, error) {

	var hotelResponse data.HotelResponse

	log.Debug().Msgf("Querying DB for hotel with ID: %d", hotelID)
	h, err := client.Hotel.Query().
		Where(hotel.IDEQ(hotelID)).
		First(ctx)
	if err != nil {
		log.Error().Err(err).Msg("There was an error querring the DB")
		return hotelResponse, err
	}

	hotelsByPrice := client.Debug().Hotel.QueryOffers(h).Order(offer.ByPrice())
	hotelOffers, err := hotelsByPrice.Offset((site - 1) * ResponseSize).Limit(ResponseSize).All(ctx)

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
			Offerid:                   singleOffer.ID,
			Hotelid:                   h.ID,
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
