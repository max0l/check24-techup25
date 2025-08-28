package testcontainer

import (
	"check24-techup25/api"
	"check24-techup25/data"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-querystring/query"
)

func TestBestOfferCacheRequest(t *testing.T) {
	apiHost, err := apiService.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get apiHost: %v", err)
	}
	apiPortMapped, err := apiService.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Fatalf("failed to get apiPortMapped port: %v", err)
	}

	queryParams := data.RequestQuerryParams{
		DepartureAirports:     []string{"BER", "FRA", "MUC", "DUS"},
		EarliestDepartureDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		LatestReturnDate:      time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		CountAdults:           2,
		CountChildren:         0,
		Duration:              7,
		Site:                  1,
	}

	values, err := query.Values(queryParams)
	if err != nil {
		t.Fatalf("failed to build query parameters: %v", err)
	}

	url := fmt.Sprintf("http://%s:%s/api/offers?%s", apiHost, apiPortMapped.Port(), values.Encode())

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("http get failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want %d. "+
			"Body: %s", resp.StatusCode, http.StatusOK, string(body))
	}

	cacheKey := api.CreateRedisCacheKeyOffers(queryParams, queryParams.Site)

	_, err = redisClient.Exists(ctx, cacheKey).Result()
	if err != nil {
		t.Fatalf("cachekey has not been set")
	}

}

func TestHotelCacheRequest(t *testing.T) {
	apiHost, err := apiService.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get apiHost: %v", err)
	}
	apiPortMapped, err := apiService.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Fatalf("failed to get apiPortMapped port: %v", err)
	}

	hotelid := 7953

	queryParams := data.RequestQuerryParams{
		DepartureAirports:     []string{"BER", "FRA", "MUC", "DUS"},
		EarliestDepartureDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		LatestReturnDate:      time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		CountAdults:           2,
		CountChildren:         0,
		Duration:              7,
		Site:                  1,
	}

	values, err := query.Values(queryParams)
	if err != nil {
		t.Fatalf("failed to build query parameters: %v", err)
	}

	url := fmt.Sprintf("http://%s:%s/api/hotel/%d?%s", apiHost, apiPortMapped.Port(), hotelid, values.Encode())

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("http get failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want %d. "+
			"Body: %s", resp.StatusCode, http.StatusOK, string(body))
	}

	cacheKey := api.CreateRedisCacheKeyOffers(queryParams, queryParams.Site)

	_, err = redisClient.Exists(ctx, cacheKey).Result()
	if err != nil {
		t.Fatalf("cachekey has not been set")
	}

}
