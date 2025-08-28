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
	"k8s.io/utils/ptr"
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

	baseParams := data.RequestQuerryParams{
		DepartureAirports:     []string{"BER", "FRA", "MUC", "DUS"},
		EarliestDepartureDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		LatestReturnDate:      time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		CountAdults:           2,
		CountChildren:         0,
		Duration:              7,
		Site:                  1,
	}

	type expectations struct {
		status             int
		countOverallOffers *int
		countOffers        *int
		expectSortedAsc    bool
		expectCheapest     *float32
		expectFirstHotel   *string
	}

	type testCases struct {
		name   string
		tweak  func(p *data.RequestQuerryParams)
		expect expectations
	}

	wantOverallOffers := 67
	wantCountSite1 := 50
	wantCountSite2 := 17
	cheapestOnSite2 := float32(2292.0)

	tests := []testCases{
		{
			name:  "site=1 baseline returns 67 and sorted by price asc",
			tweak: func(p *data.RequestQuerryParams) {}, // use base
			expect: expectations{
				status:             http.StatusOK,
				countOverallOffers: &wantOverallOffers,
				countOffers:        &wantCountSite1,
				expectSortedAsc:    true,
			},
		},
		{
			name: "site=2 returns 17, sorted asc, cheapest = 963",
			tweak: func(p *data.RequestQuerryParams) {
				p.Site = 2
			},
			expect: expectations{
				status:             http.StatusOK,
				countOverallOffers: &wantOverallOffers,
				countOffers:        &wantCountSite2,
				expectSortedAsc:    true,
				expectCheapest:     &cheapestOnSite2,
			},
		},
		{
			name: "400 when a required param is zero (CountAdults=0)",
			tweak: func(p *data.RequestQuerryParams) {
				p.CountAdults = 0
			},
			expect: expectations{
				status: http.StatusBadRequest,
			},
		},
		{
			name: "400 when Duration=0",
			tweak: func(p *data.RequestQuerryParams) {
				p.Duration = 0
			},
			expect: expectations{
				status: http.StatusBadRequest,
			},
		},
		{
			name: "404 when no offer exists (unknown airport)",
			tweak: func(p *data.RequestQuerryParams) {
				p.DepartureAirports = []string{"XXX"}
			},
			expect: expectations{
				status: http.StatusNotFound,
			},
		},
		{
			name: "404 when date window impossible (LatestReturnDate before EarliestDepartureDate)",
			tweak: func(p *data.RequestQuerryParams) {
				p.LatestReturnDate = time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
				p.EarliestDepartureDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
				p.Duration = 100
				p.CountChildren = 100
				p.CountAdults = 100
			},
			expect: expectations{
				status: http.StatusNotFound,
			},
		},
		{
			name: "boundry dates are included",
			tweak: func(p *data.RequestQuerryParams) {
				{
					p.DepartureAirports = []string{"HAJ"}
					p.EarliestDepartureDate = time.Date(2025, 9, 17, 22, 0, 0, 0, time.UTC)
					p.LatestReturnDate = time.Date(2025, 9, 24, 21, 55, 0, 0, time.UTC)
					p.CountAdults = 2
					p.CountChildren = 2
					p.Duration = 6
				}
			},
			expect: expectations{
				status:             http.StatusOK,
				countOverallOffers: ptr.To(1),
				countOffers:        ptr.To(1),
				expectSortedAsc:    true,
				expectCheapest:     ptr.To[float32](2459.0),
			},
		},
		{
			name: "400 empty airports list",
			tweak: func(p *data.RequestQuerryParams) {
				p.DepartureAirports = []string{}
			},
			expect: expectations{
				status: http.StatusBadRequest,
			},
		},
		{
			name: "404 extreme site request",
			tweak: func(p *data.RequestQuerryParams) {
				p.Site = 99999
			},
			expect: expectations{
				status: http.StatusNotFound,
			},
		},
		{
			name: "multiple but same airports",
			tweak: func(p *data.RequestQuerryParams) {
				{
					p.DepartureAirports = []string{"HAJ", "HAJ", "HAJ"}
					p.EarliestDepartureDate = time.Date(2025, 9, 17, 22, 0, 0, 0, time.UTC)
					p.LatestReturnDate = time.Date(2025, 9, 24, 21, 55, 0, 0, time.UTC)
					p.CountAdults = 2
					p.CountChildren = 2
					p.Duration = 6
				}
			},
			expect: expectations{
				status:             http.StatusOK,
				countOverallOffers: ptr.To(1),
				countOffers:        ptr.To(1),
				expectSortedAsc:    true,
				expectCheapest:     ptr.To[float32](2459.0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := baseParams
			tt.tweak(&params)

			values, err := query.Values(params)
			if err != nil {
				t.Fatalf("failed to build query parameters: %v", err)
			}

			url := fmt.Sprintf("http://%s:%s/api/offers?%s", host, mapped.Port(), values.Encode())

			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("http get failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("reading body failed: %v", err)
			}

			if resp.StatusCode != tt.expect.status {
				t.Fatalf("unexpected status: got %d, want %d. "+
					"Body: %s", resp.StatusCode, tt.expect.status, string(body))
			}

			// If not 200, nothing else to assert on this case.
			if resp.StatusCode != http.StatusOK {
				return
			}

			var result data.BestOfferPerHotelResponse
			if err := json.Unmarshal(body, &result); err != nil {
				t.Fatalf("failed to unmarshal response: %v\nBody: %s", err, string(body))
			}

			if tt.expect.countOverallOffers != nil {
				if result.CountOffers != *tt.expect.countOverallOffers {
					t.Fatalf("unexpected Overall CountOffers: got %d, "+
						"want %d", result.CountOffers, *tt.expect.countOverallOffers)
				}
			}

			if tt.expect.countOffers != nil {
				if len(result.Offers) != *tt.expect.countOffers {
					t.Fatalf("offers length mismatch: len=%d "+
						"CountOffers=%d", len(result.Offers), result.CountOffers)
				}
			}

			if tt.expect.expectSortedAsc {
				if !isPriceSortedAsc(result.Offers) {
					t.Fatalf("offers are not sorted ascending by price")
				}
			}

			if tt.expect.expectCheapest != nil {
				if len(result.Offers) == 0 {
					t.Fatalf("expected offers but got none")
				}
				got := result.Offers[0].Price
				if got != *tt.expect.expectCheapest {
					t.Fatalf("unexpected cheapest price: got %.2f, want %.2f", got, *tt.expect.expectCheapest)
				}
			}
		})
	}
}

func isPriceSortedAsc(offers []data.BestOfferPerHotel) bool {
	if len(offers) < 2 {
		return true
	}
	for i := 1; i < len(offers); i++ {
		if offers[i-1].Price > offers[i].Price {
			return false
		}
	}
	return true
}
