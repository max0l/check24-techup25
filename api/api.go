package api

import (
	_ "check24-techup25/api/docs"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:generate swag init -g api.go --parseInternal=true
//go:generate swag fmt -g api.go

type offerQuerry struct {
	DepartureAirports     []string  `binding:"required"`
	EarliestDepartureDate time.Time `binding:"required"`
	LatestReturnDate      time.Time `binding:"required"`
	CountAdults           int       `binding:"required"`
	CountChildren         int       `binding:"required"`
	Duration              int       `binding:"required"`
}

// Start starts the API
// Swagger UI will be available at: http://localhost:8080/swagger/index.html
//
//	@title	Techup25
//	@host	localhost:8080
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
//	@Param			DepartureAirports		query	[]string	true	"List of departure airports, e.g. FRA, MUC"
//	@Param			EarliestDepartureDate	query	string		true	"Earliest departure date in format RFC3339"
//	@Param			LatestReturnDate		query	string		true	"Latest return date in format RFC3339"
//	@Param			CountAdults				query	int			true	"Number of adults"
//	@Param			CountChildren			query	int			true	"Number of children"
//	@Param			Duration				query	int			true	"Duration of the trip in days"
//	@Success		200
//	@Failure		400
func handleOfferRequest(c *gin.Context) {
	var request offerQuerry
	err := c.BindQuery(&request)
	if err != nil {
		return
	} else {
		log.Info().Msgf("Got request for:\n\tdepartureAirports: %v, "+
			"\n\tearliestDepartureDate: %v, "+
			"\n\tlatestReturnDate %v,"+
			"\n\tcountAdults: %v, "+
			"\n\tcountChildren: %v, "+
			"\n\tduration: %v", request.DepartureAirports, request.EarliestDepartureDate,
			request.LatestReturnDate, request.CountAdults, request.CountChildren, request.Duration)
	}
}
