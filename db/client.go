package db

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"check24-techup25/db/ent/hotel"
	"check24-techup25/db/ent/offer"
	"check24-techup25/db/ent/predicate"
	"context"
	"errors"

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
func QueryOffers(request data.RequestQuerryParams, airports []string, client *ent.Client, ctx context.Context) []*ent.Offer {
	results, err := queryOffersFromDb(request, airports, client, ctx)
	if err != nil {
		return results
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

func queryOffersFromDb(request data.RequestQuerryParams, airports []string, client *ent.Client, ctx context.Context) ([]*ent.Offer, error) {
	results, err := client.Offer.Query().
		Where(offer.OutbounddepartureairportIn(airports...),
			offer.OutbounddeparturedatetimeGTE(request.EarliestDepartureDate),
			offer.InboundarrivaldatetimeLTE(request.LatestReturnDate),
			offer.CountadultsEQ(request.CountAdults),
			offer.CountchildrenEQ(request.CountChildren),
			offer.DurationInDaysEQ(request.Duration)).
		Order(offer.ByPrice()).WithHotel().
		All(ctx)

	log.Debug().Msgf("Querying DB for offers done, got %v results", len(results))
	if err != nil {
		log.Warn().Err(err).Msg("Could not query DB")
		return []*ent.Offer{}, err
	}
	return results, nil
}

func QueryOffersOfAHotel(h *ent.Hotel, request data.RequestQuerryParams,
	site int, client *ent.Client, ctx context.Context) (data.HotelResponse, error) {

	var hotelResponse data.HotelResponse

	hotelResponse.Hotelid = h.ID

	log.Debug().Msgf("Querying hotel with id %v", hotelResponse.Hotelid)

	hotelResponse.Hotelname = h.Hotelname
	hotelResponse.Hotelstars = h.Hotelstars

	predicates, err := makeHotelOfferPredicates(request)

	hotelsOfferQuerry := client.Hotel.QueryOffers(h).Where(predicates...)

	amount, err := getOfferAmountOfOneHotel(h, hotelsOfferQuerry, ctx)
	if err != nil {
		log.Error().Err(err).Msg("There was an error querring the DB for offers")
		return hotelResponse, err
	}

	hotelResponse.CountAvailableOffers = amount
	hotelResponse.Offers = make([]data.Offer, 0, ResponseSize)

	response, err := querryAllHotelOffersFromDb(h, site, hotelsOfferQuerry, ctx, &hotelResponse)
	if err != nil {
		return response, err
	}

	return hotelResponse, nil
}

func makeHotelOfferPredicates(request data.RequestQuerryParams) ([]predicate.Offer, error) {
	var predicates []predicate.Offer

	if request.DepartureAirports != nil && len(request.DepartureAirports) > 0 {
		predicates = append(predicates, offer.OutbounddepartureairportIn(request.DepartureAirports...))
	}

	if !request.EarliestDepartureDate.IsZero() {
		predicates = append(predicates, offer.OutbounddeparturedatetimeGTE(request.EarliestDepartureDate))
	}
	if !request.LatestReturnDate.IsZero() {
		predicates = append(predicates, offer.InboundarrivaldatetimeLTE(request.LatestReturnDate))
	}
	if request.CountAdults > 0 {
		predicates = append(predicates, offer.CountadultsEQ(request.CountAdults))
	}
	if request.CountChildren > 0 {
		predicates = append(predicates, offer.CountchildrenEQ(request.CountChildren))
	}
	if request.Duration > 0 {
		predicates = append(predicates, offer.DurationInDaysEQ(request.Duration))
	}
	if len(predicates) == 0 {
		return nil, errors.New("no predicates provided for hotel offer query")
	}

	return predicates, nil
}

func querryAllHotelOffersFromDb(h *ent.Hotel, site int, hotelsOfferQuerry *ent.OfferQuery, ctx context.Context,
	hotelResponse *data.HotelResponse) (data.HotelResponse, error) {
	hotelOffers, err := hotelsOfferQuerry.Order(offer.ByPrice()).Offset((site - 1) * ResponseSize).Limit(ResponseSize).All(ctx)

	if err != nil {
		log.Error().Err(err).Msg("There was an error querring the DB")
		return *hotelResponse, err
	}

	log.Info().Msgf("Found %d offers for hotel with ID: %d", len(hotelOffers), h.ID)

	for i, singleOffer := range hotelOffers {
		o := createOfferFromDb(h, singleOffer)
		log.Debug().Msgf("Processing offer %d for hotel %d: %+v", i+1, h.ID, o)
		hotelResponse.Offers = append(hotelResponse.Offers, o)
	}
	return *hotelResponse, nil
}

func createOfferFromDb(h *ent.Hotel, singleOffer *ent.Offer) data.Offer {
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
	return o
}

func getOfferAmountOfOneHotel(h *ent.Hotel, hotelsOfferQuerry *ent.OfferQuery, ctx context.Context) (int, error) {
	amount, err := hotelsOfferQuerry.Count(ctx)
	if err != nil {
		log.Error().Err(err).Msg("There was an error querring the DB")
		return 0, err
	}
	log.Debug().Msgf("Found %d offers for hotel with ID: %d", amount, h.ID)
	if amount == 0 {
		log.Warn().Msgf("No offers found for hotel with ID: %d", h.ID)
		return 0, errors.New("no offers found for hotel")
	}
	return amount, nil
}

func QueryHotel(hotelID int, client *ent.Client, ctx context.Context) (*ent.Hotel, error) {
	log.Debug().Msgf("Querying DB for hotel with ID: %d", hotelID)
	h, err := client.Hotel.Query().
		Where(hotel.IDEQ(hotelID)).
		First(ctx)
	if err != nil {
		log.Error().Err(err).Msg("There was an error querring the DB")
		return nil, err
	}
	return h, err
}
