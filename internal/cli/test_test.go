package cli

import (
	"context"
	"testing"

	"github.com/Hayao0819/nth/internal/app"
	diagnosticfeature "github.com/Hayao0819/nth/internal/features/diagnostic"
	"github.com/Hayao0819/nth/internal/services/auth"
)

func TestTestCommandRunsDiagnosticsWithoutStartingClient(t *testing.T) {
	t.Parallel()

	credentials := &fakeCredentials{settings: auth.Settings{Method: auth.MethodAPIToken, Token: "api-token"}}
	diagnosed := false
	deps := dependencies{
		credentials: credentials,
		setup: func(context.Context, auth.Settings, []auth.Profile, auth.OAuthStartFunc, func(auth.Settings) error) (bool, error) {
			t.Fatal("setup ran despite complete credentials")

			return false, nil
		},
		start: func(context.Context, app.API, app.Options) error {
			t.Fatal("the main client started")

			return nil
		},
		diagnose: func(_ context.Context, api diagnosticfeature.API) error {
			diagnosed = true
			if api == nil {
				t.Fatal("diagnostic API is nil")
			}

			return nil
		},
	}
	command := newCommandWith("test", deps)
	command.SetArgs([]string{"test"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !diagnosed || credentials.refreshes != 0 {
		t.Fatalf("diagnosed = %v, browser refreshes = %d", diagnosed, credentials.refreshes)
	}
}
