package auth

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/browserutils/kooky"
)

func TestCookieHeaderUsesOnlyCookiesSentToNorthAPI(t *testing.T) {
	t.Parallel()

	cookies := []*kooky.Cookie{
		{Cookie: http.Cookie{Name: "root", Value: "one", Domain: ".north.rip", Path: "/", Secure: true}},
		{Cookie: http.Cookie{Name: "api", Value: "two", Domain: "north.rip", Path: "/api", Secure: true}},
		{Cookie: http.Cookie{Name: "settings", Value: "three", Domain: "north.rip", Path: "/settings", Secure: true}},
		{Cookie: http.Cookie{Name: "other", Value: "four", Domain: "notnorth.rip", Path: "/", Secure: true}},
	}

	header, err := cookieHeader(cookies)
	if err != nil {
		t.Fatal(err)
	}
	request := &http.Request{Header: http.Header{"Cookie": []string{header}}}
	got := make(map[string]string)
	for _, cookie := range request.Cookies() {
		got[cookie.Name] = cookie.Value
	}
	if got["root"] != "one" || got["api"] != "two" {
		t.Fatalf("cookie header = %q", header)
	}
	if _, exists := got["settings"]; exists {
		t.Fatalf("unrelated path was included: %q", header)
	}
	if _, exists := got["other"]; exists {
		t.Fatalf("unrelated domain was included: %q", header)
	}
}

func TestProfileLabels(t *testing.T) {
	t.Parallel()

	profile := Profile{Browser: "chrome", Name: "Profile 1", Default: true}
	if got := profile.Label(); got != "Google Chrome · Profile 1 (default)" {
		t.Fatalf("label = %q", got)
	}
	if !profile.Same(Profile{Browser: "Chrome", Name: "Profile 1"}) {
		t.Fatal("browser matching should not be case-sensitive")
	}
	if profile.Same(Profile{Browser: "chrome", Name: "Profile 2"}) {
		t.Fatal("different profiles matched")
	}
	if profile.Same(Profile{Browser: "chrome", Name: "Profile 1", Path: "/another/Cookies"}) {
		t.Fatal("different cookie stores matched")
	}
}

func TestCookieHeaderRejectsAnEmptySet(t *testing.T) {
	t.Parallel()

	_, err := cookieHeader(nil)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %v", err)
	}
}

func TestMatchingStoresHonorsTheSelectedProfilePath(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	primaryPath := filepath.Join(directory, "primary.sqlite")
	workPath := filepath.Join(directory, "work.sqlite")
	for _, path := range []string{primaryPath, workPath} {
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	primary := &stubCookieStore{profile: Profile{
		Browser: "firefox", Name: "default", Path: primaryPath, Default: true,
	}}
	work := &stubCookieStore{profile: Profile{
		Browser: "firefox", Name: "default", Path: workPath,
	}}
	stores := []kooky.CookieStore{work, primary}

	matches := matchingStores(stores, work.profile)
	if len(matches) != 1 || matches[0].FilePath() != work.profile.Path {
		t.Fatalf("exact matches = %#v", matches)
	}
	matches = matchingStores(stores, Profile{Browser: "firefox"})
	if len(matches) != 2 || matches[0].FilePath() != primary.profile.Path {
		t.Fatalf("default matches = %#v", matches)
	}
	matches = matchingStores(stores, Profile{
		Browser: "firefox", Name: "default", Path: filepath.Join(directory, "old.sqlite"),
	})
	if len(matches) != 2 {
		t.Fatalf("moved-store matches = %#v", matches)
	}
}

type stubCookieStore struct {
	profile Profile
}

func (s *stubCookieStore) Browser() string                   { return s.profile.Browser }
func (s *stubCookieStore) Profile() string                   { return s.profile.Name }
func (s *stubCookieStore) FilePath() string                  { return s.profile.Path }
func (s *stubCookieStore) IsDefaultProfile() bool            { return s.profile.Default }
func (*stubCookieStore) Cookies(*url.URL) []*http.Cookie     { return nil }
func (*stubCookieStore) SetCookies(*url.URL, []*http.Cookie) {}
func (*stubCookieStore) SubJar(context.Context, ...kooky.Filter) (http.CookieJar, error) {
	return nil, nil
}
func (*stubCookieStore) TraverseCookies(...kooky.Filter) kooky.CookieSeq {
	return func(func(*kooky.Cookie, error) bool) {}
}
func (*stubCookieStore) Close() error { return nil }
