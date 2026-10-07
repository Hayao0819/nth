package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestStartOAuthUsesDeviceAuthorization(t *testing.T) {
	t.Parallel()

	var form url.Values
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		form = request.Form
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"device_code":"device","user_code":"ABCD-EFGH","verification_uri":"https://north.rip/device","expires_in":600,"interval":5}`))
	}))
	defer server.Close()

	manager := &Manager{
		vault:    newMemoryVault(),
		browsers: &fakeBrowserSource{},
		getenv: func(name string) string {
			if name == clientIDEnvironment {
				return "client-id"
			}

			return ""
		},
		oauthConfig: func(clientID string) *oauth2.Config {
			config := newOAuthConfig(clientID)
			config.Endpoint.DeviceAuthURL = server.URL

			return config
		},
	}
	session, err := manager.StartOAuth(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	prompt := session.Prompt()
	if prompt.VerificationURI != "https://north.rip/device" || prompt.UserCode != "ABCD-EFGH" {
		t.Fatalf("OAuth prompt = %#v", prompt)
	}
	if form.Get("client_id") != "client-id" || form.Get("client_secret") != "" {
		t.Fatalf("OAuth client form = %#v", form)
	}
	for _, scope := range []string{
		"posts.read",
		"posts.write",
		"follows.write",
		"moderation.read",
		"moderation.write",
		"bookmarks.write",
		"lists.read",
		"media.write",
		"notifications.read",
		"dm.read",
		"dm.write",
	} {
		if !strings.Contains(form.Get("scope"), scope) {
			t.Errorf("OAuth scope %q is missing from %q", scope, form.Get("scope"))
		}
	}
}

func TestOAuthClientID(t *testing.T) {
	t.Parallel()

	manager := &Manager{getenv: func(string) string { return "" }}
	if got := manager.oauthClientID(); got != defaultOAuthClientID {
		t.Fatalf("embedded client ID = %q", got)
	}
	manager.getenv = func(name string) string {
		if name == clientIDEnvironment {
			return " override-client "
		}

		return ""
	}
	if got := manager.oauthClientID(); got != "override-client" {
		t.Fatalf("overridden client ID = %q", got)
	}
}

func TestOAuthCredentialReportsNewScopes(t *testing.T) {
	t.Parallel()

	required := RequiredOAuthScopes()
	credential := OAuthCredential{
		ClientID: "client-id",
		Token:    oauth2.Token{AccessToken: "access-token"},
		Scopes:   append([]string(nil), required[:len(required)-2]...),
	}
	if missing := credential.MissingScopes(); !slices.Equal(missing, required[len(required)-2:]) {
		t.Fatalf("missing scopes = %#v", missing)
	}
	credential.Scopes = required
	if credential.NeedsAuthorization() {
		t.Fatalf("current credential needs authorization: %#v", credential.MissingScopes())
	}
}

func TestGrantedScopesUsesTokenResponse(t *testing.T) {
	t.Parallel()

	token := (&oauth2.Token{AccessToken: "access-token"}).WithExtra(map[string]any{
		"scope": "users.read posts.read",
	})
	if got := grantedScopes(token, []string{"fallback"}); !slices.Equal(got, []string{"users.read", "posts.read"}) {
		t.Fatalf("granted scopes = %#v", got)
	}
	if got := grantedScopes(&oauth2.Token{}, []string{"users.read"}); !slices.Equal(got, []string{"users.read"}) {
		t.Fatalf("fallback scopes = %#v", got)
	}
}
