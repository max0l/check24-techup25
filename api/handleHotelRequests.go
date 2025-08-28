package api

import (
	"check24-techup25/data"
	"check24-techup25/db"
	"check24-techup25/db/ent"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// handleHotelRequest will receive a hotel ID, site and a list of other optional parameters
// and returns the hotel information with its offers.
//
//	@Summary		handles hotel requests
//	@Description	handles hotel requests and returns the hotel information.
//	@Router			/api/hotel/{hotelid} [get]
//	@Param			hotelid					path		int							true	"ID of the hotel"
//	@Param			site					query		int							true	"Site, e.g. 1 or 2"
//	@Param			departureAirports		query		[]string					false	"List of departure airports, e.g. FRA, MUC"
//	@Param			earliestDepartureDate	query		string						false	"Earliest departure date in format RFC3339"
//	@Param			latestReturnDate		query		string						false	"Latest return date in format RFC3339"
//	@Param			countAdults				query		int							true	"Number of adults"
//	@Param			countChildren			query		int							false	"Number of children"
//	@Param			duration				query		int							false	"Duration of the trip in nights"
//	@Success		200						{object}	data.HotelOffersResponse	"Hotel information"
//	@Failure		400						"Bad Request"
//	@Failure		404						"Hotel Not Found"
func handleHotelRequest(c *gin.Context, client *ent.Client, redisClient *redis.Client, ctx context.Context,
	queryQueue *db.QurryQueue) {

	hotelID, err := parseHotelID(c)
	if err != nil {
		return
	}

	request := data.RequestQuerryParams{}
	log.Trace().Msgf("Got request: %v", c.Request)

	err2 := bindOptionalHotelParams(c, &request)
	if err2 != nil {
		log.Error().Err(err2).Msg("Could not bind query parameters to request")
		c.JSON(400, gin.H{"message": "Invalid query parameters"})
		return
	}

	log.Debug().Msgf("Received request for hotel with ID: %d and params: %v", hotelID, request)

	cacheKey := CreateRedisCacheKeyHotel(hotelID, request, request.Site)

	for queryQueue.Contains(cacheKey) {
		time.Sleep(1 * time.Second)
	}

	if cacheHitCheckHotel(c, cacheKey, redisClient, ctx) {
		return
	}
	queryQueue.Enqueue(cacheKey)

	hotelDb, err := queryDbForHotel(c, client, err, hotelID, ctx)
	if err != nil {
		return
	}

	hotelResponse, err := getOffersOfHotel(c, client, err, hotelDb, request, request.Site, ctx)
	if err != nil {
		return
	}

	log.Debug().Msgf("Found hotel: %v", hotelResponse)
	c.JSON(200, hotelResponse)

	queryQueue.Dequeue(cacheKey)

	go storeHotelRequestInRedis(redisClient, hotelID, request, request.Site, client, ctx, hotelResponse, queryQueue)
	return
}

// bindOptionalHotelParams binds the optional hotel params to the data.RequestQuerryParams struct.
// Since there are also optional params in the request and binding them sets them to 0, we need to
// create a new struct, which does only contain the sent params.
func bindOptionalHotelParams(c *gin.Context, params *data.RequestQuerryParams) error {
	var request data.RequestQuerryParams
	err := c.ShouldBindQuery(&request)
	if err != nil {
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

	if !hotelQueryParamsAreValid(request) {
		return errors.New("invalid hotel query parameters")
	}

	if !request.EarliestDepartureDate.IsZero() {
		params.EarliestDepartureDate = request.EarliestDepartureDate
	}
	if !request.LatestReturnDate.IsZero() {
		params.LatestReturnDate = request.LatestReturnDate
	}
	if request.CountAdults > 0 {
		params.CountAdults = request.CountAdults
	}
	if request.CountChildren >= 0 {
		params.CountChildren = request.CountChildren
	}
	if request.Duration > 0 {
		params.Duration = request.Duration
	}

	airports, err := parseAirportsToArray(request.DepartureAirports)
	if err != nil {
		log.Warn().Err(err).Msg("Could not parse airports from request")
	}
	if err == nil && len(airports) > 0 {
		params.DepartureAirports = append(params.DepartureAirports, airports...)
	}
	if request.Site > 0 {
		params.Site = request.Site
	} else {
		params.Site = 1
		log.Warn().Msgf("Site not provided, defaulting to site 1")
	}
	return nil
}

func hotelQueryParamsAreValid(request data.RequestQuerryParams) bool {
	if request.CountAdults <= 0 || request.CountChildren < 0 || request.Duration < 0 ||
		(request.CountAdults <= 0 && request.CountChildren > 0) {
		log.Error().Msg("received invalid query range")
		return false
	}
	return true
}

// parseHotelID extracts and validates hotel ID and site from the request context
func parseHotelID(c *gin.Context) (int, error) {
	hotelID, err := strconv.Atoi(c.Param("hotelid"))
	if err != nil {
		log.Error().Err(err).Msg("Could not parse hotel ID from request")
		c.JSON(400, gin.H{"message": "Invalid hotel ID"})
		return 0, err
	}
	return hotelID, nil
}

func getOffersOfHotel(c *gin.Context, client *ent.Client, err error, hotelDb *ent.Hotel,
	request data.RequestQuerryParams, site int, ctx context.Context) (data.HotelOffersResponse, error) {

	hotelResponse, err := db.QueryOffersOfAHotel(hotelDb, request, site, client, ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error querying hotel offers from DB")
		c.JSON(500, gin.H{"message": "Internal server error"})
		return data.HotelOffersResponse{}, err
	}
	return hotelResponse, nil
}

func queryDbForHotel(c *gin.Context, client *ent.Client, err error, hotelID int, ctx context.Context) (*ent.Hotel, error) {
	hotelDb, err := db.QueryHotel(hotelID, client, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Debug().Msgf("Hotel with ID %d not found in DB", hotelID)
			c.JSON(404, gin.H{"message": "Hotel not found"})
			return nil, err
		}
		log.Error().Err(err).Msg("Error querying hotel from DB")
		c.JSON(500, gin.H{"message": "Internal server error"})
		return nil, err
	}
	return hotelDb, nil
}

func cacheHitCheckHotel(c *gin.Context, cacheKey string, redisClient *redis.Client, ctx context.Context) bool {

	log.Debug().Msgf("Checking Redis cache for key: %s", cacheKey)
	cachedResponse, err := redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		log.Debug().Msgf("Found cached response for key: %s", cacheKey)
		var cachedHotel data.HotelOffersResponse
		err = json.Unmarshal(cachedResponse, &cachedHotel)
		if err == nil {
			log.Debug().Msgf("Returning cached hotel: %v", cachedHotel)
			c.JSON(200, cachedHotel)
			return true
		} else {
			log.Error().Err(err).Msg("Could not unmarshal cached hotel response")
			c.JSON(500, gin.H{"message": "Internal server error"})
			return false
		}
	} else {
		log.Debug().Msgf("Cache Key not found: %s", cacheKey)
	}
	return false
}
