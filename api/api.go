package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

type offerQuerry struct {
	departureAirports     []string  `binding:"required"`
	earliestDepartureDate time.Time `binding:"required"`
	latestReturnDate      time.Time `binding:"required"`
	countAdults           int       `binding:"required"`
	countChildren         int       `binding:"required"`
	duration              int       `binding:"required"`
}

func Start(port int) {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.GET("/offers", handleOfferRequest)

	log.Info().Msgf("Starting Server on port %v", port)
	err := router.Run(":" + strconv.Itoa(port))
	if err != nil {
		log.Error().Msgf("Could not start server on port %v", port)
		return
	}
}

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
