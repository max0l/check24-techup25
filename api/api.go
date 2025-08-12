package api

import (
	_ "check24-techup25/api/docs"
	"check24-techup25/data"
	"check24-techup25/db"
	"check24-techup25/db/ent"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Start starts the API
// Swagger UI will be available at: http://localhost:8080/swagger/index.html
//
//	@title	Techup25
//	@host	localhost:8080
func Start(port int, client *ent.Client, redisClient *redis.Client, ctx context.Context) {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/offers", func(c *gin.Context) {
		handleOfferRequest(c, client, redisClient, ctx)
	})

	router.GET("/hotel/:hotelid", func(c *gin.Context) {
		handleHotelRequest(c, client, redisClient, ctx)
	})

	log.Info().Msgf("Starting Server on port %v", port)
	err := router.Run(":" + strconv.Itoa(port))
	if err != nil {
		log.Error().Msgf("Could not start server on port %v", port)
		return
	}
}

// handleHotelRequest will receive a hotel ID and returns the hotel information
//
//	@Summary		handles hotel requests
//	@Description	handles hotel requests and returns the hotel information.
//	@Router			/hotel/{hotelid} [get]
//	@Param			site	query	int			true				"Site, e.g. 1 or 2"
//	@Param			hotelid	path	int			true				"ID of the hotel"
//	@Success		200		body	{object}	data.HotelResponse	"Hotel information"
//	@Failure		400		"Bad Request"
//	@Failure		404		"Hotel Not Found"
func handleHotelRequest(c *gin.Context, client *ent.Client, redisClient *redis.Client, ctx context.Context) {
	hotelID, err := strconv.Atoi(c.Param("hotelid"))
	if err != nil {
		log.Error().Err(err).Msg("Could not parse hotel ID from request")
		c.JSON(400, gin.H{"message": "Invalid hotel ID"})
		return
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
		return
	}

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
			return
		}
	}

	log.Debug().Msgf("Received request for hotel with ID: %d and site %v", hotelID, site)
	hotelResponse, err := db.QueryHotel(hotelID, site, client, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(404, gin.H{"message": "Hotel not found"})
			return
		}
		log.Error().Err(err).Msg("Error querying hotel from DB")
		c.JSON(500, gin.H{"message": "Internal server error"})
	}

	go storeHotelRequestInRedis(redisClient, hotelID, site, client, ctx)

	log.Debug().Msgf("Found hotel: %v", hotelResponse)
	c.JSON(200, hotelResponse)
}

// handleOfferRequest will receive a query with different parameters and returns a list of offers
//
//	@Summary		handles offer requests
//	@Description	handles offer requests and returns a list of offers.
//	@Router			/offers [get]
//	@Param			DepartureAirports		query		[]string				true	"List of departure airports, e.g. FRA, MUC"
//	@Param			EarliestDepartureDate	query		string					true	"Earliest departure date in format RFC3339"
//	@Param			LatestReturnDate		query		string					true	"Latest return date in format RFC3339"
//	@Param			CountAdults				query		int						true	"Number of adults"
//	@Param			CountChildren			query		int						false	"Number of children"
//	@Param			Duration				query		int						true	"Duration of the trip in days"
//	@Param			Site					query		int						true	"Site, e.g. 1 or 2"
//	@Success		200						{object}	[]data.OfferResponse	"List of offers matching the request"
//	@Failure		400						"Bad Request"
//	@Failure		404						"No Offers Found"
func handleOfferRequest(ginc *gin.Context, client *ent.Client, redisClient *redis.Client, ctx context.Context) {
	var request data.OfferQuerry
	err := ginc.BindQuery(&request)
	if err != nil {
		log.Error().Err(err).Msg("Could not bind query parameters")
		return
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

	if request.EarliestDepartureDate.After(request.LatestReturnDate) {
		log.Error().Msg("Earliest departure date is after latest return date")
		ginc.JSON(400, gin.H{"message": "Earliest departure date cannot be after latest return date"})
		return
	}

	if request.Duration < 1 {
		log.Warn().Msg("Duration cant be less than a 1 day")
		ginc.JSON(400, gin.H{"message": "Duration must be at least 1 day"})
		return
	}

	airports, err := parseAirportsToArray(request.DepartureAirports)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse airports from request")
		ginc.JSON(400, gin.H{"message": "Invalid Airport(s)"})
		return
	}

	cacheKey := CreateRedisCacheKeyOffers(request, request.Site)
	log.Debug().Msgf("Checking Redis cache for key: %s", cacheKey)
	cachedResponse, err := redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		log.Debug().Msgf("Found cached response for key: %s", cacheKey)
		var cachedOffers []data.OfferResponse
		err = json.Unmarshal(cachedResponse, &cachedOffers)
		if err == nil {
			log.Debug().Msgf("Returning cached offers: %v", cachedOffers)
			ginc.JSON(200, cachedOffers)
			log.Info().Msg("Request handled successfully with cached response")
			return
		}
	}

	//TODO: Introduce page size
	log.Debug().Msg("Querying DB for offers")
	response := db.QueryOffers(request, airports, client, ctx)
	if len(response) == 0 {
		log.Warn().Msg("No offers found for the given request")
		ginc.JSON(404, gin.H{"message": "No offers found"})
		return
	} else {
		log.Info().Msgf("Found %v offers for the given request", len(response))
	}

	offerResponse := make([]data.OfferResponse, 0, 50)

	parseOfferDbResponse(response, &offerResponse, request)

	go storeOfferRequestInRedis(redisClient, request, response, ctx)

	ginc.JSON(200, offerResponse)
	log.Debug().Msgf("Response: %v", offerResponse)
	log.Info().Msg("Request handled successfully")
}

func parseOfferDbResponse(uniqueOffers []*ent.Offer, responsesStore *[]data.OfferResponse, request data.OfferQuerry) {

	processedResults := 0
	for i := db.ResponseSize * (request.Site - 1); i < len(uniqueOffers); i++ {

		if processedResults >= db.ResponseSize {
			log.Debug().Msg("Reached response size limit, stopping processing offers")
			break
		}
		processedResults++

		offerResult := uniqueOffers[i]
		hotelResult := uniqueOffers[i].Edges.Hotel

		log.Debug().Msgf("Processing offer: %v", i)

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
		*responsesStore = append(*responsesStore, response)
	}
}

func parseAirportsToArray(airports []string) ([]string, error) {
	chunks := strings.Split(airports[0], ",")
	if len(chunks) == 0 {
		log.Warn().Msg("Could not split airports into chunks, returning empty array")
		return []string{}, fmt.Errorf("invalid Airport(s)")
	}
	var apList []string
	for i, chunk := range chunks {
		log.Debug().Msgf("Evaluating chunk %d: '%s'", i, chunk)
		if fitsAirPortNamingCon(chunk) && chunk != "" {
			log.Debug().Msgf("Chunk '%d' is a valid airport code: '%s'", i, chunk)
			apList = append(apList, chunk)
		}
	}
	log.Debug().Msgf("Parsed airports from %v (len %v) to %v (len: %v)", airports, len(airports), apList, len(apList))
	return apList, nil
}

func fitsAirPortNamingCon(s string) bool {
	re := regexp.MustCompile("(?i)[A-Z][A-Z][A-Z]")
	return re.MatchString(s)
}
