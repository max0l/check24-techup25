package testcontainer

import (
	"check24-techup25/api"
	"check24-techup25/db"
	"check24-techup25/db/ent"
	"check24-techup25/logging"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	dc          *compose.DockerCompose
	ctx         context.Context
	apiService  *testcontainers.DockerContainer
	dbClient    *ent.Client
	redisClient *redis.Client
)

func TestMain(m *testing.M) {
	logging.SetLogger()

	tempComposeFileLocation := filepath.Join(os.TempDir(), "docker-compose-test.yml")

	log.Info().Msgf("Temp Docker Compose: %v", tempComposeFileLocation)

	ctx = context.Background()

	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projectRoot := filepath.Dir(workingDir)

	log.Debug().Msgf("Project Root dir: %v", projectRoot)

	composePath := filepath.Join(projectRoot, "docker-compose.yml")
	if _, err := os.Stat(composePath); err != nil {
		panic(err)
	}

	newComposeFile := modifyNewComposeForTesting(composePath, tempComposeFileLocation, projectRoot)

	dc, err = compose.NewDockerCompose(newComposeFile)

	if err := dc.WaitForService("api", wait.ForHealthCheck()).
		WaitForService("seed_data", wait.ForExit().WithExitTimeout(10*time.Second)).
		WithEnv(map[string]string{
			"OFFERS_CSV_PATH": filepath.Join(projectRoot, "testCSVs/small/offer.csv"),
			"REDIS_URL":       "redis:6379",
			"REDIS_PASSWORD":  "password",
			"REDIS_DB":        "0",
		}).
		Up(ctx, compose.Wait(true)); err != nil {
		CleanUpCompose(dc, ctx, tempComposeFileLocation)
		panic(err)
	}

	apiService, err = dc.ServiceContainer(ctx, "api")
	if err != nil {
		log.Error().Err(err).Msgf("Could not find api container")
		CleanUpCompose(dc, ctx, tempComposeFileLocation)
		os.Exit(1)
	}

	_, err = dc.ServiceContainer(ctx, "db")
	if err != nil {
		log.Error().Err(err).Msgf("Could not find db container")
		CleanUpCompose(dc, ctx, tempComposeFileLocation)
		os.Exit(1)
	}

	_, err = dc.ServiceContainer(ctx, "redis")
	if err != nil {
		log.Error().Err(err).Msgf("Could not find redis container")
		CleanUpCompose(dc, ctx, tempComposeFileLocation)
		os.Exit(1)
	}

	redisClient, err = api.GetRedisClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to Redis")
		CleanUpCompose(dc, ctx, tempComposeFileLocation)
		os.Exit(1)
	}

	dbDSN := "host=localhost port=5432 user=user dbname=check24 password=password sslmode=disable"

	dbClient, err = db.NewClient(dbDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("DSN is not valid")
		CleanUpCompose(dc, ctx, tempComposeFileLocation)
		os.Exit(1)
	}

	code := m.Run()

	api.CloseRedisClient(redisClient)

	CleanUpCompose(dc, ctx, tempComposeFileLocation)

	os.Exit(code)
}
