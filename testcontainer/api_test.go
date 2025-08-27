package testcontainer

import (
	"fmt"
	"io"
	"net/http"
	"testing"
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
