package db

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"check24-techup25/db/ent/hotel"
	"check24-techup25/db/ent/offer"
	"check24-techup25/db/ent/predicate"
	"context"
	"entgo.io/ent/dialect/sql"
	"errors"
	"sync"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// QurryQueue is a simple queue implementation for tracking database queries.
type QurryQueue struct {
	sync.RWMutex
	queue map[string]bool
}

// NewDbQueue creates a new QurryQueue instance.
func NewDbQueue() *QurryQueue {
	return &QurryQueue{
		queue: make(map[string]bool),
	}
}

// Enqueue adds a new query to the queue.
func (qq *QurryQueue) Enqueue(key string) {
	qq.Lock()
	defer qq.Unlock()
	qq.queue[key] = true
}

// Dequeue removes a query from the queue.
func (qq *QurryQueue) Dequeue(key string) {
	qq.Lock()
	defer qq.Unlock()
	if qq.queue[key] {
		delete(qq.queue, key)
	}
}

// Contains checks if a query is in the queue.
func (qq *QurryQueue) Contains(key string) bool {
	qq.RLock()
	defer qq.RUnlock()
	return qq.queue[key]
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

// Migrate applies the schema migrations to the database.
func Migrate(client ent.Client) error {
	if err := client.Schema.Create(context.Background()); err != nil {
		return err
	}
	return nil
}

// QueryOffers retrieves offers based on the provided request parameters and airports.
// It filters out duplicates based on hotel ID, counts the duplicates and returns the unique offers.
func QueryOffers(request data.RequestQuerryParams, airports []string, client *ent.Client,
	ctx context.Context) ([]*ent.Offer, map[int]int, error) {
	results, err := queryOffersFromDb(request, airports, client, ctx)
	if err != nil {
		return []*ent.Offer{}, nil, err
	}

	uniqueOffers := make([]*ent.Offer, 0, len(results))
	seenHotels := make(map[int]int)

	for _, result := range results {
		offerHotelId := result.Edges.Hotel.ID
		if seenHotels[offerHotelId] == 0 {
			uniqueOffers = append(uniqueOffers, result)
			seenHotels[offerHotelId]++
		} else {
			seenHotels[offerHotelId]++
		}
	}

	log.Debug().Msgf("Found %v unique offers", len(uniqueOffers))

	// this checks if the requested site is even possible to display (if there enough unique offers)
	if ResponseSize*(request.Site-1) >= len(uniqueOffers) {
		log.Warn().Msg("Response size limit is larger than available offers")
		return []*ent.Offer{}, nil, errors.New("response size limit is larger than available offers")
	}

	log.Debug().Msg("Queried offers")
	return uniqueOffers, seenHotels, nil
}

// queryOffersFromDb queries offers from the database based on the given request parameters.
// It does that by getting all offers that match the request parameters and airports ordered by price.
func queryOffersFromDb(request data.RequestQuerryParams, airports []string, client *ent.Client, ctx context.Context) ([]*ent.Offer, error) {
	results, err := client.Debug().Offer.
		Query().
		Where(
			offer.OutbounddepartureairportIn(airports...),
			offer.OutbounddeparturedatetimeGTE(request.EarliestDepartureDate),
			offer.InboundarrivaldatetimeLTE(request.LatestReturnDate),
			offer.CountadultsEQ(request.CountAdults),
			offer.CountchildrenEQ(request.CountChildren),
			offer.DurationInDaysEQ(request.Duration),
		).
		Modify(func(s *sql.Selector) {

			t := offer.Table        // "offers"
			hk := offer.HotelColumn // e.g. "hotelid" (inverse edge FK)
			pk := offer.FieldPrice  // "price"

			qTbl := `"` + t + `"`            // "offers"
			qHotel := qTbl + `."` + hk + `"` // "offers"."hotelid"
			qPrice := qTbl + `."` + pk + `"` // "offers"."price"

			s.SelectExpr(sql.Expr(
				`DISTINCT ON (` + qHotel + `) ` + qTbl + `.*`,
			))

			s.OrderBy(
				qHotel,
				qPrice+` ASC`,
			)

			s.Limit(100)
		}).
		WithHotel().
		All(ctx)

	log.Debug().Msgf("Querying DB for offers done, got %v results", len(results))
	if err != nil {
		log.Warn().Err(err).Msg("Could not query DB")
		return []*ent.Offer{}, err
	}
	return results, nil
}

// QueryOffersOfAHotel queries offers for a specific hotel.
func QueryOffersOfAHotel(h *ent.Hotel, request data.RequestQuerryParams,
	site int, client *ent.Client, ctx context.Context) (data.HotelOffersResponse, error) {

	var hotelResponse data.HotelOffersResponse

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

// makeHotelOfferPredicates creates a slice of predicates based on the request parameters.
// These predicates are used to filter the hotel offers in the database.
// Most of them are optional and only added if the corresponding request parameter is not empty or zero.
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

// querryAllHotelOffersFromDb queries all hotel offers from the database based on the provided predicates and site.
// Other then QueryOffers, not all offers of a hotel are returned. Here the Limit and Site is used in the DB query right away.
func querryAllHotelOffersFromDb(h *ent.Hotel, site int, hotelsOfferQuerry *ent.OfferQuery, ctx context.Context,
	hotelResponse *data.HotelOffersResponse) (data.HotelOffersResponse, error) {
	hotelOffers, err := hotelsOfferQuerry.Order(offer.ByPrice()).Offset((site - 1) * ResponseSize).Limit(ResponseSize).All(ctx)

	if err != nil {
		log.Error().Err(err).Msg("There was an error querring the DB")
		return *hotelResponse, err
	}

	log.Info().Msgf("Found %d offers for hotel with ID: %d", len(hotelOffers), h.ID)

	for i, singleOffer := range hotelOffers {
		o := createOfferFromDb(h, singleOffer)
		log.Trace().Msgf("Processing offer %d for hotel %d: %+v", i+1, h.ID, o)
		hotelResponse.Offers = append(hotelResponse.Offers, o)
	}
	return *hotelResponse, nil
}

// createOfferFromDb creates a data.Offer object from an ent.Offer.
func createOfferFromDb(h *ent.Hotel, singleOffer *ent.Offer) data.Offer {
	o := data.Offer{
		Offerid:                   singleOffer.ID,
		Hotelid:                   h.ID,
		DurationInDays:            singleOffer.DurationInDays,
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

// getOfferAmountOfOneHotel returns the number of offers for a given hotel.
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

// QueryHotel queries the database for a hotel with a given ID.
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
