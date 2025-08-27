package testcontainer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

func CleanUpCompose(dc *compose.DockerCompose, ctx context.Context, tempComposeFileLocation string) {
	log.Info().Msg("Removing Temp Compose File")
	if err := os.Remove(tempComposeFileLocation); err != nil {
		log.Warn().Err(err).Msg("Could not remove temp docker compose")
	}
	log.Info().Msg("Cleaning up Docker Compose resources")
	err := dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveVolumes(true), compose.RemoveImagesLocal)
	if err != nil {
		log.Warn().Err(err).Msg("Could not put down docker compose")
	}
}

func modifyNewComposeForTesting(composeFile string, testComposePath string, projectRootPath string) string {
	copyFile(composeFile, testComposePath)
	changeComposeFileForTesting(testComposePath, projectRootPath)

	return testComposePath
}

func changeComposeFileForTesting(newComposePath string, projectRoot string) {
	data, err := os.ReadFile(newComposePath)
	if err != nil {
		panic(err)
	}

	newComposeContext := fmt.Sprintf("context: %v", projectRoot)
	hotelCSVPath := filepath.Join(projectRoot, "data/hotels.csv")
	offersCSVPath := filepath.Join(projectRoot, "testCSVs/small/offer.csv")

	replacements := map[string]string{
		"./data/hotels.csv":            hotelCSVPath,
		"../offers/offer.csv":          offersCSVPath,
		"context: .":                   newComposeContext,
		"${SEED_DATA:-true}":           "true",
		"${SKIP_INVALID_HOTELS:-true}": "false",
	}
	for old, newContents := range replacements {
		data = bytes.ReplaceAll(data, []byte(old), []byte(newContents))
	}

	fmt.Printf("New Contents: %v", string(data))

	if err := os.WriteFile(newComposePath, data, 0o644); err != nil {
		panic(err)
	}
}

func copyFile(src string, dst string) {
	sourceFile, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		panic(err)
	}
	defer destinationFile.Close()

	// Copy the content
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		panic(err)
	}

	// Flush file metadata to disk
	err = destinationFile.Sync()
	if err != nil {
		panic(err)
	}
}
