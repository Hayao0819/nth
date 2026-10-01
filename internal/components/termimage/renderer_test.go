package termimage

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

func TestRendererLoadsAndPlacesImage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "image/png")
		picture := image.NewRGBA(image.Rect(0, 0, 2, 2))
		picture.Set(0, 0, color.RGBA{R: 255, A: 255})
		if err := png.Encode(writer, picture); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	renderer := &Renderer{enabled: true, kitty: true, client: server.Client(), entries: make(map[string]entry)}
	command := renderer.Load(context.Background(), server.URL, 4, 2)
	if command == nil {
		t.Fatal("image load did not start")
	}
	rawCommand, handled := renderer.Update(command())
	if !handled || rawCommand == nil {
		t.Fatal("loaded image was not handled")
	}
	raw, ok := rawCommand().(tea.RawMsg)
	if !ok || !strings.Contains(raw.Msg.(string), "\x1b_G") {
		t.Fatalf("raw command = %#v", raw)
	}

	view := renderer.View(server.URL, 4, 2)
	lines := strings.Split(view, "\n")
	if len(lines) != 2 || ansi.StringWidth(lines[0]) != 4 || !strings.ContainsRune(view, kitty.Placeholder) {
		t.Fatalf("placeholder = %q", view)
	}
	if duplicate := renderer.Load(context.Background(), server.URL, 4, 2); duplicate != nil {
		t.Fatal("ready image was downloaded again")
	}
}

func TestSupportedRecognizesKitty(t *testing.T) {
	t.Parallel()

	values := map[string]string{"TERM": "xterm-kitty"}
	if !Supported(func(name string) string { return values[name] }) {
		t.Fatal("xterm-kitty was not recognized")
	}
	if Supported(func(name string) string {
		if name == "TERM" {
			return "xterm-256color"
		}

		return ""
	}) {
		t.Fatal("an unrelated terminal was recognized as Kitty")
	}
}

func TestNewEnablesFallbackOutsideKitty(t *testing.T) {
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TMUX", "")

	renderer := New(true)
	if !renderer.Enabled() {
		t.Fatal("image previews were disabled outside Kitty")
	}
	if renderer.kitty {
		t.Fatal("an unrelated terminal selected the Kitty renderer")
	}
}

func TestRendererFallsBackToHalfBlocks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "image/png")
		picture := image.NewRGBA(image.Rect(0, 0, 4, 4))
		for y := range 4 {
			for x := range 4 {
				picture.Set(x, y, color.RGBA{R: uint8(x * 60), G: uint8(y * 60), B: 120, A: 255})
			}
		}
		if err := png.Encode(writer, picture); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	renderer := &Renderer{enabled: true, client: server.Client(), entries: make(map[string]entry)}
	command := renderer.Load(context.Background(), server.URL, 4, 2)
	if command == nil {
		t.Fatal("image load did not start")
	}
	rawCommand, handled := renderer.Update(command())
	if !handled || rawCommand != nil {
		t.Fatalf("half-block image update = (%#v, %t)", rawCommand, handled)
	}

	view := renderer.View(server.URL, 4, 2)
	lines := strings.Split(view, "\n")
	if len(lines) != 2 {
		t.Fatalf("image has %d lines, want 2: %q", len(lines), view)
	}
	for _, line := range lines {
		if width := ansi.StringWidth(line); width != 4 {
			t.Fatalf("line width = %d, want 4: %q", width, line)
		}
	}
	if !strings.ContainsRune(view, '▀') || strings.ContainsRune(view, kitty.Placeholder) {
		t.Fatalf("half-block image = %q", view)
	}
}

func TestResolveURLAcceptsNorthMediaPaths(t *testing.T) {
	t.Parallel()

	resolved, err := resolveURL("/media/2026/10/photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := resolved.String(), "https://north.rip/media/2026/10/photo.jpg"; got != want {
		t.Fatalf("resolved URL = %q, want %q", got, want)
	}
	for _, value := range []string{"media/photo.jpg", "//example.com/photo.jpg", "file:///tmp/photo.jpg"} {
		if _, err := resolveURL(value); err == nil {
			t.Errorf("resolveURL(%q) succeeded", value)
		}
	}
}

func TestFetchResolvesNorthMediaPath(t *testing.T) {
	t.Parallel()

	var picture bytes.Buffer
	if err := png.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got, want := request.URL.String(), "https://north.rip/media/avatar.png"; got != want {
			t.Errorf("request URL = %q, want %q", got, want)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(bytes.NewReader(picture.Bytes())),
			Header:     make(http.Header),
		}, nil
	})}
	renderer := &Renderer{enabled: true, kitty: true, client: client, entries: make(map[string]entry)}
	sequence, view, err := renderer.fetch(context.Background(), "/media/avatar.png", 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if view != "" {
		t.Fatalf("fallback view = %q", view)
	}
	if !strings.Contains(sequence, "\x1b_G") {
		t.Fatalf("graphics sequence = %q", sequence)
	}
}

func TestFetchWrapsGraphicsForTmux(t *testing.T) {
	t.Parallel()

	var picture bytes.Buffer
	if err := png.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(bytes.NewReader(picture.Bytes())),
			Header:     make(http.Header),
		}, nil
	})}
	renderer := &Renderer{enabled: true, kitty: true, tmux: true, client: client, entries: make(map[string]entry)}
	sequence, view, err := renderer.fetch(context.Background(), "https://north.rip/media/avatar.png", 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if view != "" {
		t.Fatalf("fallback view = %q", view)
	}
	if !strings.Contains(sequence, "\x1bPtmux;") {
		t.Fatalf("graphics sequence was not wrapped for tmux: %q", sequence)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestFitPreservesAspectRatio(t *testing.T) {
	t.Parallel()

	source := image.NewRGBA(image.Rect(0, 0, 1200, 800))
	resized := fit(source, 600, 240)
	if resized.Bounds().Dx() != 360 || resized.Bounds().Dy() != 240 {
		t.Fatalf("resized image = %v", resized.Bounds())
	}
}
