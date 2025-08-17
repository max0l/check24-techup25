package api

import (
	"check24-techup25/data"
	"check24-techup25/db"
	"check24-techup25/db/ent"
	"context"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// handleHotelRequest will receive a hotel ID and returns the hotel information
//
//	@Summary		handles hotel requests
//	@Description	handles hotel requests and returns the hotel information.
//	@Router			/hotel/{hotelid} [get]
//	@Param			hotelid					path		int					true	"ID of the hotel"
//	@Param			site					query		int					true	"Site, e.g. 1 or 2"
//	@Param			departureAirports		query		[]string			false	"List of departure airports, e.g. FRA, MUC"
//	@Param			earliestDepartureDate	query		string				false	"Earliest departure date in format RFC3339"
//	@Param			latestReturnDate		query		string				false	"Latest return date in format RFC3339"
//	@Param			countAdults				query		int					false	"Number of adults"
//	@Param			countChildren			query		int					false	"Number of children"
//	@Param			duration				query		int					false	"Duration of the trip in days"
//	@Success		200						{object}	data.HotelResponse	"Hotel information"
//	@Failure		400						"Bad Request"
//	@Failure		404						"Hotel Not Found"
func handleHotelRequest(c *gin.Context, client *ent.Client, redisClient *redis.Client, ctx context.Context) {

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

	if cacheHitCheckHotel(c, hotelID, request, request.Site, err, redisClient, ctx) {
		return
	}

	hotelDb, err := queryDbForHotel(c, client, err, hotelID, ctx)
	if err != nil {
		return
	}

	hotelResponse, err := getOffersOfHotel(c, client, err, hotelDb, request, request.Site, ctx)
	if err != nil {
		return
	}

	go storeHotelRequestInRedis(redisClient, hotelID, request, request.Site, client, ctx)

	log.Debug().Msgf("Found hotel: %v", hotelResponse)
	c.JSON(200, hotelResponse)
	return
}

func bindOptionalHotelParams(c *gin.Context, params *data.RequestQuerryParams) error {
	var request data.RequestQuerryParams
	err := c.ShouldBindQuery(&request)
	if err != nil {
		log.Error().Err(err).Msg("Could not bind query parameters to request")
		c.JSON(400, gin.H{"message": "Invalid query parameters"})
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
	if !request.EarliestDepartureDate.IsZero() {
		params.EarliestDepartureDate = request.EarliestDepartureDate
	}
	if !request.LatestReturnDate.IsZero() {
		params.LatestReturnDate = request.LatestReturnDate
	}
	if request.CountAdults > 0 {
		params.CountAdults = request.CountAdults
	}
	if request.CountChildren > 0 {
		params.CountChildren = request.CountChildren
	}
	if request.Duration > 0 {
		params.Duration = request.Duration
	}

	airports, err := parseAirportsToArray(request.DepartureAirports)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse airports from request")
		return err
	}
	if airports != nil && len(airports) > 0 {
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
	request data.RequestQuerryParams, site int, ctx context.Context) (data.HotelResponse, error) {

	hotelResponse, err := db.QueryOffersOfAHotel(hotelDb, request, site, client, ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error querying hotel offers from DB")
		c.JSON(500, gin.H{"message": "Internal server error"})
		return data.HotelResponse{}, err
	}
	return hotelResponse, nil
}

func queryDbForHotel(c *gin.Context, client *ent.Client, err error, hotelID int, ctx context.Context) (*ent.Hotel, error) {
	hotelDb, err := db.QueryHotel(hotelID, client, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"message": "Hotel not found"})
			return nil, err
		}
		log.Error().Err(err).Msg("Error querying hotel from DB")
		c.JSON(500, gin.H{"message": "Internal server error"})
		return nil, err
	}
	return hotelDb, nil
}

func cacheHitCheckHotel(c *gin.Context, hotelID int, request data.RequestQuerryParams,
	site int, err error, redisClient *redis.Client, ctx context.Context) bool {

	cacheKey := CreateRedisCacheKeyHotel(hotelID, request, site)
	log.Debug().Msgf("Checking Redis cache for key: %s", cacheKey)
	cachedResponse, err := redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		log.Debug().Msgf("Found cached response for key: %s", cacheKey)
		var cachedHotel data.HotelResponse
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
		log.Debug().Msgf("Error getting cached response for key: %s", cacheKey)
	}
	return false
}
