//go:build integration

package diagnostic

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/services/northapi"
)

func TestIntegrationReadOnlyChecks(t *testing.T) {
	token := os.Getenv("NORTH_API_KEY")
	if token == "" {
		t.Skip("NORTH_API_KEY is not set")
	}
	client, err := north.NewClient(token, north.WithUserAgent("nth/integration-test"))
	if err != nil {
		t.Fatal(err)
	}
	api := northapi.NewHybrid(client, nil, "", nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for _, check := range checks(api) {
		t.Run(check.name, func(t *testing.T) {
			if _, err := check.run(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}
