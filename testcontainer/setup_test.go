package testcontainer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	dc           *compose.DockerCompose
	ctx          context.Context
	apiService   *testcontainers.DockerContainer
	dbService    *testcontainers.DockerContainer
	redisService *testcontainers.DockerContainer
)

func TestMain(m *testing.M) {

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

	dbService, err = dc.ServiceContainer(ctx, "db")
	if err != nil {
		log.Error().Err(err).Msgf("Could not find db container")
		CleanUpCompose(dc, ctx, tempComposeFileLocation)
		os.Exit(1)
	}

	redisService, err = dc.ServiceContainer(ctx, "redis")
	if err != nil {
		log.Error().Err(err).Msgf("Could not find redis container")
		CleanUpCompose(dc, ctx, tempComposeFileLocation)
		os.Exit(1)
	}

	code := m.Run()

	CleanUpCompose(dc, ctx, tempComposeFileLocation)

	os.Exit(code)

}
