package api

import (
	"check24-techup25/data"
	"check24-techup25/db"
	"check24-techup25/db/ent"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func GetRedisClient(ctx context.Context) (*redis.Client, error) {
	redisUrl, redisPassword, redisDB := GetRedisEnvs()
	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Ping the Redis server to check if the connection is successful
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func GetRedisEnvs() (string, string, int) {
	redisUrl := "localhost:6379"
	redisPassword := ""
	redisDB := 0

	if url := os.Getenv("REDIS_URL"); url != "" {
		redisUrl = url
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		redisPassword = password
	}
	if rdb := os.Getenv("REDIS_DB"); rdb != "" {
		redisDB, _ = strconv.Atoi(rdb)
	}
	log.Debug().Msgf("REDIS_URL: %s, REDIS_PASSWORD: %s, REDIS_DB: %d", redisUrl, redisPassword, redisDB)
	return redisUrl, redisPassword, redisDB
}

func CloseRedisClient(client *redis.Client) {
	err := client.Close()
	if err != nil {
		log.Warn().Err(err).Msg("redis client close error")
	} else {
		log.Info().Msg("redis client closed successfully")
	}
}

func storeHotelRequestInRedis(redisClient *redis.Client, hotelId int, request data.RequestQuerryParams,
	requestSite int, entClient *ent.Client, ctx context.Context, hotelResponse data.HotelResponse) {
	site := requestSite

	createRequestedSiteKey := CreateRedisCacheKeyHotel(hotelId, request, request.Site)

	if !cacheKeyAlreadyExists(redisClient, createRequestedSiteKey, ctx) {
		jsonHotel, err := json.Marshal(hotelResponse)
		if err != nil {
			log.Error().Err(err).Msg("Could not marshal hotel response to JSON")
		}
		err = redisClient.Set(ctx, createRequestedSiteKey, jsonHotel, time.Minute*20).Err()
		if err != nil {
			log.Warn().Err(err).Msg("redis store hotel error")
		}
		log.Debug().Msgf("Stored hotel %d for site %d successfully in Redis with key: %s", hotelId, request.Site, createRequestedSiteKey)
	} else {
		log.Debug().Msgf("Hotel %d for site %d already exists in Redis with key: %s", hotelId, request.Site, createRequestedSiteKey)
	}
	site++

	// Only cache 10 sites for each request because some results could contain over 100000 offers (which means 2000 sites)
	for site < requestSite+10 {
		redisKey := CreateRedisCacheKeyHotel(hotelId, request, site)

		if !cacheKeyAlreadyExists(redisClient, createRequestedSiteKey, ctx) {
			log.Debug().Msgf("Cache key already exists: %s", redisKey)
			site++
			continue
		}

		hotelDb, err := db.QueryHotel(hotelId, entClient, ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				log.Debug().Msgf("Hotel with ID %d not found in DB", hotelId)
				return
			}
			log.Error().Err(err).Msg("Error querying hotel from DB")
			return
		}

		hotelResponse, err := db.QueryOffersOfAHotel(hotelDb, request, site, entClient, ctx)
		if err != nil {
			log.Error().Err(err).Msg("Error querying hotel offers from DB")
			return
		}

		if len(hotelResponse.Offers) == 0 {
			return
		}

		jsonHotel, err := json.Marshal(hotelResponse)
		if err != nil {
			log.Error().Err(err).Msg("Could not marshal hotel response to JSON")
			site++
			continue
		}
		err = redisClient.Set(ctx, redisKey, jsonHotel, time.Minute*20).Err()
		if err != nil {
			log.Warn().Err(err).Msg("redis store hotel error")
			site++
			continue
		}
		site++
		log.Debug().Msgf("Stored hotel %d for site %d successfully in Redis with key: %s", hotelId, site, redisKey)
	}
}

func cacheKeyAlreadyExists(client *redis.Client, key string, ctx context.Context) bool {
	result, err := client.Exists(ctx, key).Result()
	if err != nil || result == 0 {
		return false
	}
	return true
}

func storeOfferRequestInRedis(redisClient *redis.Client, request data.RequestQuerryParams, uniqueOffers []*ent.Offer, ctx context.Context) {
	var modifiedRequest = request
	if request.Site != 1 {
		modifiedRequest.Site = 1
	}
	for {
		redisKey := CreateRedisCacheKeyOffers(modifiedRequest, modifiedRequest.Site)

		responseStore := data.OfferResponse{
			Offers:      make([]data.OfferWithHotel, 0, 50),
			CountOffers: len(uniqueOffers),
		}
		parseOfferDbResponse(uniqueOffers, &responseStore, modifiedRequest)
		if len(responseStore.Offers) == 0 {
			return
		}

		modifiedRequest.Site++
		jsonOffer, err := json.Marshal(responseStore)
		if err != nil {
			log.Error().Err(err).Msg("Could not marshal offers to JSON")
			continue
		}
		err = redisClient.Set(ctx, redisKey, jsonOffer, time.Minute*20).Err()
		if err != nil {
			log.Warn().Err(err).Msg("redis store offer error")
			continue
		}
		log.Debug().Msgf("store offer %s successfully", redisKey)
	}

}

func CreateRedisCacheKeyOffers(request data.RequestQuerryParams, site int) string {
	return fmt.Sprintf("offers:%s:%s:%s:%d:%d:%d:%d",
		strings.Join(request.DepartureAirports, ","), request.EarliestDepartureDate.String(), request.LatestReturnDate.String(),
		request.CountAdults, request.CountChildren, request.Duration, site)
}

func CreateRedisCacheKeyHotel(hotelId int, request data.RequestQuerryParams, site int) string {
	// Airports
	airports := "any"
	if len(request.DepartureAirports) > 0 {
		airports = strings.Join(request.DepartureAirports, ",")
	}

	// Dates
	earliest := "none"
	if !request.EarliestDepartureDate.IsZero() {
		earliest = request.EarliestDepartureDate.String()
	}

	latest := "none"
	if !request.LatestReturnDate.IsZero() {
		latest = request.LatestReturnDate.String()
	}
	adults := "none"
	if request.CountAdults > 0 {
		adults = strconv.Itoa(request.CountAdults)
	}

	children := "none"
	if request.CountChildren > 0 {
		children = strconv.Itoa(request.CountChildren)
	}

	duration := "none"
	if request.Duration > 0 {
		duration = strconv.Itoa(request.Duration)
	}

	return fmt.Sprintf(
		"hotel:%d:%s:%s:%s:%s:%s:%s:%d",
		hotelId,
		airports,
		earliest,
		latest,
		adults,
		children,
		duration,
		site,
	)
}
