package testcontainer

import (
	"context"
	"testing"
)

func TestDbCounts(t *testing.T) {

	type testCase struct {
		name           string
		run            func(context.Context) (int, error)
		expectedAmount int
	}

	tests := []testCase{
		{
			name: "Check Offer Amount in DB",
			run: func(ctx context.Context) (int, error) {
				return dbClient.Offer.Query().Count(ctx)
			},
			expectedAmount: 1099,
		},
		{
			name: "Check Hotels Amount in DB",
			run: func(ctx context.Context) (int, error) {
				return dbClient.Hotel.Query().Count(ctx)
			},
			expectedAmount: 1282,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.run(ctx)
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}
			if got != tt.expectedAmount {
				t.Errorf("count = %d; want %d", got, tt.expectedAmount)
			}
		})
	}
}
