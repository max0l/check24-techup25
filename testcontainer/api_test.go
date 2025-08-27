package testcontainer

import (
	"check24-techup25/data"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-querystring/query"
)

func TestApiPing(t *testing.T) {
	host, err := apiService.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}
	mapped, err := apiService.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	url := fmt.Sprintf("http://%s:%s/api/ping", host, mapped.Port())

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("http get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body failed: %v", err)
	}

	got := string(body)
	want := "pong"
	if got != want {
		t.Fatalf("unexpected body: got %q, want %q", got, want)
	}
}

func TestApiBestOfferPerHotel(t *testing.T) {
	host, err := apiService.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}
	mapped, err := apiService.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	requestPramsStruct := data.RequestQuerryParams{
		DepartureAirports:     []string{"BER", "FRA", "MUC", "DUS"},
		EarliestDepartureDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		LatestReturnDate:      time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		CountAdults:           2,
		CountChildren:         0,
		Duration:              7,
		Site:                  1,
	}

	parameters, err := query.Values(requestPramsStruct)
	if err != nil {
		t.Fatalf("failed to build query parameters: %v", err)
	}

	fmt.Printf("Parameters: %+v\n", parameters.Encode())

	url := fmt.Sprintf("http://%s:%s/api/offers?%v", host, mapped.Port(), parameters.Encode())

	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("http get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body failed: %v", err)
	}

	var result data.BestOfferPerHotelResponse

	err = json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	got := result.CountOffers
	want := 67
	if got != want {
		t.Fatalf("unexpected body: got %d, want %d", got, want)
	}

	//when requesting site 2, it should be 17 results
	//cheapest should be 963
}
