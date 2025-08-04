package api

import (
	_ "check24-techup25/api/docs"
	"check24-techup25/data"
	"check24-techup25/db"
	"check24-techup25/db/ent"
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:generate swag init -g api.go --parseInternal=true
//go:generate swag fmt -g api.go

// Start starts the API
// Swagger UI will be available at: http://localhost:8080/swagger/index.html
//
//	@title	Techup25
//	@host	localhost:8080
func Start(port int, client *ent.Client, ctx context.Context) {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/offers", func(c *gin.Context) {
		handleOfferRequest(c, client, ctx)
	})

	log.Info().Msgf("Starting Server on port %v", port)
	err := router.Run(":" + strconv.Itoa(port))
	if err != nil {
		log.Error().Msgf("Could not start server on port %v", port)
		return
	}
}

// handleOfferRequest will receive a query with different parameters and returns a list of offers
//
//	@Summary		handles offer requests
//	@Description	handles offer requests and returns a list of offers.
//	@Router			/offers [get]
//	@Param			DepartureAirports		query	[]string	true	"List of departure airports, e.g. FRA, MUC"
//	@Param			EarliestDepartureDate	query	string		true	"Earliest departure date in format RFC3339"
//	@Param			LatestReturnDate		query	string		true	"Latest return date in format RFC3339"
//	@Param			CountAdults				query	int			true	"Number of adults"
//	@Param			CountChildren			query	int			true	"Number of children"
//	@Param			Duration				query	int			true	"Duration of the trip in days"
//	@Param			Site					query	int			true	"Site, e.g. 1 or 2"
//	@Success		200
//	@Failure		400
func handleOfferRequest(ginc *gin.Context, client *ent.Client, ctx context.Context) {
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

	//TODO: Introduce page size
	log.Debug().Msg("Querying DB for offers")
	response := db.QueryOffers(request, client, ctx)
	if len(response) == 0 {
		log.Warn().Msg("No offers found for the given request")
		ginc.JSON(404, gin.H{"message": "No offers found"})
		return
	} else {
		log.Info().Msgf("Found %v offers for the given request", len(response))
	}
	ginc.JSON(200, response)
	log.Debug().Msgf("Response: %v", response)
	log.Info().Msg("Request handled successfully")
}
