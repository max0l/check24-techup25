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
	if db := os.Getenv("REDIS_DB"); db != "" {
		redisDB, _ = strconv.Atoi(db)
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

func storeHotelRequestInRedis(redisClient *redis.Client, hotelId int, requestSite int, entClient *ent.Client, ctx context.Context) {
	site := requestSite
	if site < 1 {
		site = 1
	}

	for {
		redisKey := CreateRedisCacheKeyHotel(hotelId, site)
		if redisKeyExists(redisClient, redisKey, ctx) {
			site++
			continue
		}

		hotelResponse, err := db.QueryHotel(hotelId, site, entClient, ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				log.Warn().Msgf("Hotel with ID %d not found", hotelId)
				return
			}
			log.Error().Err(err).Msg("Error querying hotel from DB")
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
		log.Debug().Msgf("Stored hotel %d for site %d successfully in Redis with key: %s", hotelId, site, redisKey)
	}
}

func storeOfferRequestInRedis(redisClient *redis.Client, request data.OfferQuerry, uniqueOffers []*ent.Offer, ctx context.Context) {
	var modifiedRequest = request
	if request.Site != 1 {
		modifiedRequest.Site = 1
	}
	for {
		redisKey := CreateRedisCacheKeyOffers(modifiedRequest, modifiedRequest.Site)
		log.Debug().Msgf("Checking Redis cache for key: %s", redisKey)

		if redisKeyExists(redisClient, redisKey, ctx) {
			modifiedRequest.Site++
			continue
		}

		responseStore := make([]data.OfferResponse, 0, 50)
		parseOfferDbResponse(uniqueOffers, &responseStore, modifiedRequest)
		if len(responseStore) == 0 {
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

func redisKeyExists(redisClient *redis.Client, redisKey string, ctx context.Context) bool {
	exists, err := redisClient.Exists(ctx, redisKey).Result()
	if err == nil && exists > 0 {
		log.Debug().Msgf("Found cached response for key: %s", redisKey)
		ok, err := redisClient.Expire(ctx, redisKey, time.Minute*20).Result()
		if err != nil {
			log.Error().Err(err).Msg("Could not extend Redis cache expiration")
		}
		if !ok {
			log.Warn().Msgf("Could not extend expiration for key: %s", redisKey)
		} else {
			log.Debug().Msgf("Extended expiration for key: %s", redisKey)
		}
		return true
	}
	return false

}

func CreateRedisCacheKeyOffers(request data.OfferQuerry, site int) string {
	return fmt.Sprintf("offers:%s:%s:%s:%d:%d:%d:%d",
		strings.Join(request.DepartureAirports, ","), request.EarliestDepartureDate.String(), request.EarliestDepartureDate.String(),
		request.CountAdults, request.CountChildren, request.Duration, site)
}

func CreateRedisCacheKeyHotel(hotelId int, site int) string {
	return fmt.Sprintf("hotel:%d:%d", hotelId, site)
}
