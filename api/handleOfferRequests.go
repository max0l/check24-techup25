package api

import (
	"check24-techup25/data"
	"check24-techup25/db"
	"check24-techup25/db/ent"
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// handleOfferRequest will receive a query with different parameters and returns a list of offers
//
//	@Summary		handles offer requests
//	@Description	handles offer requests and returns a list of offers.
//	@Router			/api/offers [get]
//	@Param			departureAirports		query		[]string						true	"List of departure airports, e.g. FRA, MUC"
//	@Param			earliestDepartureDate	query		string							true	"Earliest departure date in format RFC3339"
//	@Param			latestReturnDate		query		string							true	"Latest return date in format RFC3339"
//	@Param			countAdults				query		int								true	"Number of adults"
//	@Param			countChildren			query		int								false	"Number of children"
//	@Param			duration				query		int								true	"Duration of the trip in days"
//	@Param			site					query		int								false	"Site, e.g. 1 or 2"
//	@Success		200						{object}	data.BestOfferPerHotelResponse	"List of offers matching the request"
//	@Failure		400						"Bad Request"
//	@Failure		404						"No Offers Found"
func handleOfferRequest(ginc *gin.Context, client *ent.Client, redisClient *redis.Client, ctx context.Context,
	queryQueue *db.QurryQueue) {
	var request data.RequestQuerryParams
	log.Trace().Msgf("Got request: %v", ginc.Request)
	err := bindOfferRequest(ginc, &request)
	if err != nil {
		return
	}

	if checkIfParamsAreRight(ginc, request) {
		return
	}

	airports, err := parseAirportsToArray(request.DepartureAirports)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse airports from request")
		ginc.JSON(400, gin.H{"message": "Invalid Airport(s)"})
		return
	}

	cacheKey := CreateRedisCacheKeyOffers(request, request.Site)

	for queryQueue.Contains(cacheKey) {
		time.Sleep(1 * time.Second)
	}

	if cacheHitCheckOffer(ginc, cacheKey, err, redisClient, ctx) {
		return
	}

	queryQueue.Enqueue(cacheKey)

	dbResponse, amountMap, done := getOffersFromDB(ginc, client, request, airports, ctx)
	if done {
		return
	}

	offerResponse := data.BestOfferPerHotelResponse{
		Offers: make([]data.BestOfferPerHotel, 0, 50),
	}

	parseOfferDbResponse(dbResponse, amountMap, &offerResponse, request.Site)

	ginc.JSON(200, offerResponse)
	log.Debug().Msgf("Response: %v", offerResponse)
	log.Info().Msg("Request handled successfully")
	queryQueue.Dequeue(cacheKey)
	go storeOfferRequestInRedis(redisClient, request, dbResponse, amountMap, ctx)
	return
}

func getOffersFromDB(ginc *gin.Context, client *ent.Client, request data.RequestQuerryParams, airports []string,
	ctx context.Context) ([]*ent.Offer, map[int]int, bool) {
	log.Debug().Msg("Querying DB for offers")
	response, amountMap, err := db.QueryOffers(request, airports, client, ctx)
	if len(response) == 0 {
		log.Warn().Msg("No offers found for the given request")
		ginc.JSON(404, gin.H{"message": "No offers found"})
		return nil, nil, true
	} else if err != nil {
		log.Error().Err(err).Msg("Error querying DB")
		ginc.JSON(500, gin.H{"message": "Error querying DB"})
	}
	log.Info().Msgf("Found %v offers for the given request", len(response))
	return response, amountMap, false
}

func cacheHitCheckOffer(ginc *gin.Context, cacheKey string, err error, redisClient *redis.Client, ctx context.Context) bool {
	log.Debug().Msgf("Checking Redis cache for key: %s", cacheKey)
	cachedResponse, err := redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		log.Debug().Msgf("Found cached response for key: %s", cacheKey)
		var cachedOffers data.BestOfferPerHotelResponse
		err = json.Unmarshal(cachedResponse, &cachedOffers)
		if err == nil {
			log.Debug().Msgf("Returning cached offers: %v", cachedOffers)
			ginc.JSON(200, cachedOffers)
			log.Info().Msg("Request handled successfully with cached response")
			return true
		} else {
			log.Error().Err(err).Msg("Could not unmarshal cached offers response")
		}
	}
	return false
}

func bindOfferRequest(ginc *gin.Context, request *data.RequestQuerryParams) error {
	err := ginc.BindQuery(&request)
	if err != nil {
		log.Error().Err(err).Msg("Could not bind query parameters")
		ginc.JSON(400, gin.H{"message": "Invalid query parameters"})
		return err
	} else {
		log.Debug().Msgf("Got request for:\n\tdepartureAirports: %v, "+
			"\n\tearliestDepartureDate: %v, "+
			"\n\tlatestReturnDate %v,"+
			"\n\tcountAdults: %v, "+
			"\n\tcountChildren: %v, "+
			"\n\tduration: %v"+
			"\n\tsite: %v", request.DepartureAirports, request.EarliestDepartureDate,
			request.LatestReturnDate, request.CountAdults, request.CountChildren, request.Duration,
			request.Site)
	}
	return err
}

func parseOfferDbResponse(uniqueOffers []*ent.Offer, amountMap map[int]int, responseStore *data.BestOfferPerHotelResponse, site int) {

	responseStore.CountOffers = len(uniqueOffers)
	processedResults := 0
	for i := db.ResponseSize * (site - 1); i < len(uniqueOffers); i++ {

		if processedResults >= db.ResponseSize {
			log.Debug().Msg("Reached response size limit, stopping processing offers")
			break
		}
		processedResults++

		offerResult := uniqueOffers[i]
		hotelResult := uniqueOffers[i].Edges.Hotel

		log.Trace().Msgf("Processing offer: %v", i)

		offerWithHotel := parseOfferResult(hotelResult, offerResult, amountMap)
		responseStore.Offers = append(responseStore.Offers, offerWithHotel)
	}
}

func parseOfferResult(hotelResult *ent.Hotel, offerResult *ent.Offer, amountMap map[int]int) data.BestOfferPerHotel {
	offerWithHotel := data.BestOfferPerHotel{
		HotelName:                 hotelResult.Hotelname,
		HotelStars:                hotelResult.Hotelstars,
		HotelId:                   hotelResult.ID,
		DurationInDays:            offerResult.DurationInDays,
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
		OfferAmount:               amountMap[hotelResult.ID],
	}
	return offerWithHotel
}
