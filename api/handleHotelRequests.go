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
//	@Param			site	query		int					true	"Site, e.g. 1 or 2"
//	@Param			departureAirports		query		[]string				false	"List of departure airports, e.g. FRA, MUC"
//	@Param			earliestDepartureDate	query		string					false	"Earliest departure date in format RFC3339"
//	@Param			latestReturnDate		query		string					false	"Latest return date in format RFC3339"
//	@Param			countAdults				query		int						false	"Number of adults"
//	@Param			countChildren			query		int						false	"Number of children"
//	@Param			duration				query		int						false	"Duration of the trip in days"
//	@Param			site					query		int						false	"Site, e.g. 1 or 2"
//	@Param			hotelid	path		int					true	"ID of the hotel"
//	@Success		200		{object}	data.HotelResponse	"Hotel information"
//	@Failure		400		"Bad Request"
//	@Failure		404		"Hotel Not Found"
func handleHotelRequest(c *gin.Context, client *ent.Client, redisClient *redis.Client, ctx context.Context) {

	hotelID, site, err := parseHotelIDAndSite(c)
	if err != nil {
		return
	}

	log.Debug().Msgf("Received request for hotel with ID: %d and site %v", hotelID, site)

	if cacheHitCheckHotel(c, hotelID, site, err, redisClient, ctx) {
		return
	}

	hotelDb, err := queryDbForHotel(c, client, err, hotelID, ctx)
	if err != nil {
		return
	}

	hotelResponse, err := getOffersOfHotel(c, client, err, hotelDb, site, ctx)
	if err != nil {
		return
	}

	go storeHotelRequestInRedis(redisClient, hotelID, site, client, ctx)

	log.Debug().Msgf("Found hotel: %v", hotelResponse)
	c.JSON(200, hotelResponse)
	return
}

// parseHotelIDAndSite extracts and validates hotel ID and site from the request context
func parseHotelIDAndSite(c *gin.Context) (int, int, error) {
	hotelID, err := strconv.Atoi(c.Param("hotelid"))
	if err != nil {
		log.Error().Err(err).Msg("Could not parse hotel ID from request")
		c.JSON(400, gin.H{"message": "Invalid hotel ID"})
		return 0, 0, err
	}
	siteQuery, exists := c.GetQuery("site")
	if !exists {
		siteQuery = "1"
		log.Warn().Msgf("Could not find site query from request")
	}
	site, err := strconv.Atoi(siteQuery)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse site from request")
		c.JSON(400, gin.H{"message": "Invalid site"})
		return 0, 0, err
	}
	return hotelID, site, nil
}

func getOffersOfHotel(c *gin.Context, client *ent.Client, err error, hotelDb *ent.Hotel, site int, ctx context.Context) (data.HotelResponse, error) {
	hotelResponse, err := db.QueryOffersOfAHotel(hotelDb, site, client, ctx)
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

func cacheHitCheckHotel(c *gin.Context, hotelID int, site int, err error, redisClient *redis.Client, ctx context.Context) bool {
	cacheKey := CreateRedisCacheKeyHotel(hotelID, site)
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
