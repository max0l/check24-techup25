package api

import (
	_ "check24-techup25/api/docs"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"strconv"
	"time"
)

//go:generate swag init -g api.go --parseInternal=true

type offerQuerry struct {
	departureAirports     []string  `binding:"required"`
	earliestDepartureDate time.Time `binding:"required"`
	latestReturnDate      time.Time `binding:"required"`
	countAdults           int       `binding:"required"`
	countChildren         int       `binding:"required"`
	duration              int       `binding:"required"`
}

// Start starts the API
// Swagger UI will be available at: http://localhost:8080/swagger/index.html
//
//	@title			Techup25
//	@host			localhost:8080
func Start(port int) {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/offers", handleOfferRequest)

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
//	@Param          offer query offerQuerry true "offer"
//	@Success		200
//	@Failure		400
func handleOfferRequest(c *gin.Context) {
	var request offerQuerry
	err := c.BindQuery(&request)
	if err != nil {
		return
	} else {
		log.Info().Msgf("Got request for:\n\tdepartureAirports: %v, "+
			"\tearliestDepartureDate: %v, "+
			"\tlatestReturnDate %v,"+
			"\tcountAdults: %v, "+
			"\tcountChildren: %v, "+
			"\tduration: %v", request.departureAirports, request.earliestDepartureDate,
			request.latestReturnDate, request.countAdults, request.countChildren, request.duration)
	}
}
