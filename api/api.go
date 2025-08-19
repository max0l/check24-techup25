package api

import (
	_ "check24-techup25/api/docs"
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
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
	//I know setting the cors to default is not the best idea, but for this project it is sufficient
	router.Use(cors.Default())
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

func checkIfParamsAreRight(ginc *gin.Context, request data.RequestQuerryParams) bool {
	if request.EarliestDepartureDate.After(request.LatestReturnDate) {
		log.Error().Msg("Earliest departure date is after latest return date")
		ginc.JSON(400, gin.H{"message": "Earliest departure date cannot be after latest return date"})
		return true
	}

	if request.Duration < 1 {
		log.Warn().Msgf("Duration cant be less than a 1 day, but it was %v", request.Duration)
		ginc.JSON(400, gin.H{"message": "Duration must be at least 1 day"})
		return true
	}
	if request.Site < 1 {
		log.Warn().Msgf("Site cant be less than 1, but it was %v", request.Site)
		ginc.JSON(400, gin.H{"message": "Site must be at least 1"})
		return true
	}
	if request.CountAdults < 1 {
		log.Warn().Msgf("CountAdults cant be less than 1, but it was %v", request.CountAdults)
		ginc.JSON(400, gin.H{"message": "CountAdults must be at least 1"})
		return true
	}
	return false
}

func parseAirportsToArray(airports []string) ([]string, error) {
	if len(airports) == 0 || airports[0] == "" {
		log.Warn().Msg("No airports provided, returning empty array")
		return []string{}, fmt.Errorf("no Airport(s) provided")
	}

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
