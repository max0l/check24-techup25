package api

import (
	_ "check24-techup25/api/docs"
	"check24-techup25/data"
	"check24-techup25/db"
	"check24-techup25/db/ent"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Start launches the HTTP Server/REST API in Gin Release mode, sets up Recovery, CORS and Logging.
// Swagger UI will be available at: http://localhost:8080/swagger/index.html
//
//	@title	Techup25
//	@host	localhost:8080
func Start(port int, client *ent.Client, redisClient *redis.Client, ctx context.Context, queryQueue *db.QurryQueue) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(cors.Default())
	//I know setting the cors to default is not the best idea, but for this project it is sufficient
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/offers", func(c *gin.Context) {
		handleOfferRequest(c, client, redisClient, ctx, queryQueue)
	})

	router.GET("/hotel/:hotelid", func(c *gin.Context) {
		handleHotelRequest(c, client, redisClient, ctx, queryQueue)
	})

	log.Info().Msgf("Starting Server on port %v", port)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: router.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Msgf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println("Server Shutdown:", err)
	}
	fmt.Println("Server exiting")
}

// cancelIfParamsAreWrong validates incoming query parameters for a request. If any of the parameters are invalid,
// it will cancel the request with an error response code 400 and returns true.
func cancelIfParamsAreWrong(ginc *gin.Context, request data.RequestQuerryParams) bool {
	if request.EarliestDepartureDate.IsZero() || request.LatestReturnDate.IsZero() {
		log.Error().Msg("Dates cannot be empty")
		ginc.JSON(400, gin.H{"message": "Date(s) cannot be empty"})
		return true
	}

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

	if request.LatestReturnDate.Sub(request.EarliestDepartureDate).Hours() < float64(24*request.Duration) {
		log.Warn().Msgf("Selected dates are smaller than the duration")
		ginc.JSON(400, gin.H{"message": "Selected dates are smaller than the duration"})
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

// parseAirportsToArray converts a slice containing a comma‑separated string of
// airport codes into a slice of individual, validated airport codes.
func parseAirportsToArray(airports []string) ([]string, error) {
	if len(airports) == 0 || airports[0] == "" {
		log.Warn().Msg("No airports provided, returning empty array")
		return []string{}, data.ErrNoAirportProvided
	}

	chunks := strings.Split(airports[0], ",")
	if len(chunks) == 0 {
		log.Warn().Msg("Could not split airports into chunks, returning empty array")
		return []string{}, data.ErrInvalidAirport
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

// fitsAirPortNamingCon checks if a given string is an airport code.
// This is done by checking if the string matches the regex pattern [A-Z][A-Z][A-Z]
func fitsAirPortNamingCon(s string) bool {
	re := regexp.MustCompile("(?i)[A-Z][A-Z][A-Z]")
	return re.MatchString(s)
}
