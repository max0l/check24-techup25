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

func storeHotelRequestInRedis(redisClient *redis.Client, hotelId int, request data.RequestQuerryParams, requestSite int,
	entClient *ent.Client, ctx context.Context, hotelResponse data.HotelResponse, queryQueue *db.QurryQueue) {

	storeInitialHotelResponseSite(redisClient, hotelId, request, requestSite, ctx, hotelResponse)

	storeAdditionalHotelResponseSites(redisClient, hotelId, request, requestSite, entClient, ctx, queryQueue)
}

func storeInitialHotelResponseSite(redisClient *redis.Client, hotelId int, request data.RequestQuerryParams, site int,
	ctx context.Context, hotelResponse data.HotelResponse) {
	cacheKey := CreateRedisCacheKeyHotel(hotelId, request, site)

	if cacheKeyAlreadyExists(redisClient, cacheKey, ctx) {
		log.Debug().Msgf("Hotel %d for site %d already exists in Redis with key: %s", hotelId,
			request.Site, cacheKey)
		return
	}

	if err := saveHotelResponseToRedis(redisClient, cacheKey, hotelResponse, ctx); err != nil {
		log.Warn().Err(err).Msg("redis store hotel error (initial site)")
		return
	}

	log.Debug().Msgf("Stored hotel %d for site %d successfully in Redis with key: %s", hotelId, request.Site, cacheKey)
}

func storeAdditionalHotelResponseSites(redisClient *redis.Client, hotelId int, request data.RequestQuerryParams,
	requestSite int, entClient *ent.Client, ctx context.Context, queryQueue *db.QurryQueue) {

	for site := requestSite + 1; site < requestSite+10; site++ {
		cacheKey := CreateRedisCacheKeyHotel(hotelId, request, site)

		if cacheKeyAlreadyExists(redisClient, cacheKey, ctx) {
			log.Debug().Msgf("Cache key already exists: %s", cacheKey)
			continue
		}

		if queryQueue.Contains(cacheKey) {
			continue
		} else {
			queryQueue.Enqueue(cacheKey)
		}

		hotelResponse, ok := fetchHotelOffers(hotelId, request, site, entClient, ctx)
		if !ok || len(hotelResponse.Offers) == 0 {
			return
		}

		if err := saveHotelResponseToRedis(redisClient, cacheKey, hotelResponse, ctx); err != nil {
			log.Warn().Err(err).Msg("redis store hotel error (additional sites)")
			continue
		}

		queryQueue.Dequeue(cacheKey)

		log.Debug().Msgf("Stored hotel %d for site %d successfully in Redis with key: %s", hotelId, site, cacheKey)
	}
}

func fetchHotelOffers(hotelId int, request data.RequestQuerryParams, site int, entClient *ent.Client,
	ctx context.Context) (data.HotelResponse, bool) {

	hotelDb, err := db.QueryHotel(hotelId, entClient, ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Debug().Msgf("Hotel with ID %d not found in DB", hotelId)
		} else {
			log.Error().Err(err).Msg("Error querying hotel from DB")
		}
		return data.HotelResponse{}, false
	}

	hotelResponse, err := db.QueryOffersOfAHotel(hotelDb, request, site, entClient, ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error querying hotel offers from DB")
		return data.HotelResponse{}, false
	}

	return hotelResponse, true
}

func saveHotelResponseToRedis(redisClient *redis.Client, cacheKey string, hotelResponse data.HotelResponse,
	ctx context.Context) error {

	jsonHotel, err := json.Marshal(hotelResponse)
	if err != nil {
		log.Error().Err(err).Msg("Could not marshal hotel response to JSON")
		return err
	}

	return redisClient.Set(ctx, cacheKey, jsonHotel, time.Minute*20).Err()
}

func cacheKeyAlreadyExists(client *redis.Client, key string, ctx context.Context) bool {
	result, err := client.Exists(ctx, key).Result()
	if err != nil || result == 0 {
		return false
	}
	_ = client.Expire(ctx, key, time.Minute*20).Val()
	return true
}

func storeOfferRequestInRedis(redisClient *redis.Client, request data.RequestQuerryParams, uniqueOffers []*ent.Offer,
	amountMap map[int]int, ctx context.Context) {

	for site := request.Site + 1; site < request.Site+10; site++ {
		redisKey := CreateRedisCacheKeyOffers(request, site)

		if cacheKeyAlreadyExists(redisClient, redisKey, ctx) {
			log.Debug().Msgf("Cache key already exists: %s", redisKey)
			continue
		}

		responseStore := data.BestOfferPerHotelResponse{
			Offers:      make([]data.BestOfferPerHotel, 0, 50),
			CountOffers: len(uniqueOffers),
		}
		parseOfferDbResponse(uniqueOffers, amountMap, &responseStore, site)
		if len(responseStore.Offers) == 0 {
			return
		}

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
