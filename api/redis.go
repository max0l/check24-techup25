package api

import (
	"context"
	"os"
	"strconv"

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
