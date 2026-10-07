package termimage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	_ "golang.org/x/image/webp"
)

const (
	maxDownload    = 12 << 20
	northAPIOrigin = "https://api.north.rip"

	AvatarColumns = 2
	AvatarRows    = 1
)

type entry struct {
	loading bool
	ready   bool
	id      int
	view    string
}

type loadedMsg struct {
	target   *Renderer
	key      string
	id       int
	sequence string
	view     string
	err      error
}

type Renderer struct {
	enabled bool
	kitty   bool
	tmux    bool
	client  *http.Client

	mu      sync.Mutex
	entries map[string]entry
}

func New(enabled bool) *Renderer {
	return &Renderer{
		enabled: enabled,
		kitty:   Supported(os.Getenv),
		tmux:    strings.TrimSpace(os.Getenv("TMUX")) != "",
		client:  &http.Client{Timeout: 12 * time.Second},
		entries: make(map[string]entry),
	}
}

func Supported(getenv func(string) string) bool {
	if getenv == nil {
		return false
	}
	if strings.TrimSpace(getenv("KITTY_WINDOW_ID")) != "" {
		return true
	}

	return strings.Contains(strings.ToLower(getenv("TERM")), "kitty")
}

func (r *Renderer) Enabled() bool {
	return r != nil && r.enabled
}

func (r *Renderer) Load(ctx context.Context, rawURL string, columns, rows int) tea.Cmd {
	if !r.Enabled() || strings.TrimSpace(rawURL) == "" || columns <= 0 || rows <= 0 {
		return nil
	}
	key := imageKey(rawURL, columns, rows)
	r.mu.Lock()
	current, exists := r.entries[key]
	if exists && (current.loading || current.ready) {
		r.mu.Unlock()

		return nil
	}
	id := imageID(key)
	r.entries[key] = entry{loading: true, id: id}
	r.mu.Unlock()

	return func() tea.Msg {
		sequence, view, err := r.fetch(ctx, rawURL, id, columns, rows)

		return loadedMsg{target: r, key: key, id: id, sequence: sequence, view: view, err: err}
	}
}

func (r *Renderer) Update(msg tea.Msg) (tea.Cmd, bool) {
	loaded, ok := msg.(loadedMsg)
	if !ok || loaded.target != r {
		return nil, false
	}
	r.mu.Lock()
	r.entries[loaded.key] = entry{ready: loaded.err == nil, id: loaded.id, view: loaded.view}
	r.mu.Unlock()
	if loaded.err != nil || loaded.sequence == "" {
		return nil, true
	}

	return tea.Raw(loaded.sequence), true
}

func (r *Renderer) View(rawURL string, columns, rows int) string {
	if !r.Enabled() || columns <= 0 || rows <= 0 {
		return ""
	}
	key := imageKey(rawURL, columns, rows)
	r.mu.Lock()
	current := r.entries[key]
	r.mu.Unlock()
	if !current.ready {
		return ""
	}
	if current.view != "" {
		return current.view
	}

	red := current.id >> 16 & 0xff
	green := current.id >> 8 & 0xff
	blue := current.id & 0xff
	color := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", red, green, blue)
	lines := make([]string, rows)
	for row := range rows {
		var line strings.Builder
		line.WriteString(color)
		for column := range columns {
			line.WriteRune(kitty.Placeholder)
			line.WriteRune(kitty.Diacritic(row))
			line.WriteRune(kitty.Diacritic(column))
		}
		line.WriteString("\x1b[39m")
		lines[row] = line.String()
	}

	return strings.Join(lines, "\n")
}

func (r *Renderer) fetch(ctx context.Context, rawURL string, id, columns, rows int) (string, string, error) {
	parsed, err := resolveURL(rawURL)
	if err != nil {
		return "", "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", "", err
	}
	response, err := r.client.Do(request)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", "", fmt.Errorf("image request returned %s", response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxDownload+1))
	if err != nil {
		return "", "", err
	}
	if len(data) > maxDownload {
		return "", "", errors.New("image is too large")
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", "", err
	}
	pixels := int64(config.Width) * int64(config.Height)
	if config.Width <= 0 || config.Height <= 0 ||
		config.Width > 8192 || config.Height > 8192 || pixels > 32_000_000 {
		return "", "", errors.New("image dimensions are too large")
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", "", err
	}
	if !r.kitty {
		return "", renderHalfBlocks(decoded, columns, rows), nil
	}
	decoded = fit(decoded, max(1, columns*12), max(1, rows*24))
	var sequence bytes.Buffer
	var chunkFormatter func(string) string
	if r.tmux {
		chunkFormatter = ansi.TmuxPassthrough
	}
	err = kitty.EncodeGraphics(&sequence, decoded, &kitty.Options{
		Action:           kitty.TransmitAndPut,
		Format:           kitty.PNG,
		ID:               id,
		Columns:          columns,
		Rows:             rows,
		VirtualPlacement: true,
		Quiet:            2,
		Chunk:            true,
		ChunkFormatter:   chunkFormatter,
	})
	if err != nil {
		return "", "", err
	}

	return sequence.String(), "", nil
}

func renderHalfBlocks(source image.Image, columns, rows int) string {
	if columns <= 0 || rows <= 0 {
		return ""
	}
	maximumHeight := rows * 2
	scaled := fit(source, columns, maximumHeight)
	bounds := scaled.Bounds()
	left := (columns - bounds.Dx()) / 2
	top := (maximumHeight - bounds.Dy()) / 2
	lines := make([]string, rows)
	for row := range rows {
		var line strings.Builder
		for column := range columns {
			upper := halfBlockPixel(scaled, column-left, row*2-top)
			lower := halfBlockPixel(scaled, column-left, row*2+1-top)
			switch {
			case upper.alpha == 0 && lower.alpha == 0:
				line.WriteString("\x1b[39;49m ")
			case lower.alpha == 0:
				fmt.Fprintf(&line, "\x1b[38;2;%d;%d;%d;49m▀", upper.red, upper.green, upper.blue)
			case upper.alpha == 0:
				fmt.Fprintf(&line, "\x1b[38;2;%d;%d;%d;49m▄", lower.red, lower.green, lower.blue)
			default:
				fmt.Fprintf(
					&line,
					"\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm▀",
					upper.red, upper.green, upper.blue,
					lower.red, lower.green, lower.blue,
				)
			}
		}
		line.WriteString("\x1b[39;49m")
		lines[row] = line.String()
	}

	return strings.Join(lines, "\n")
}

type halfBlockColor struct {
	red   uint8
	green uint8
	blue  uint8
	alpha uint8
}

func halfBlockPixel(source image.Image, x, y int) halfBlockColor {
	bounds := source.Bounds()
	if x < 0 || y < 0 || x >= bounds.Dx() || y >= bounds.Dy() {
		return halfBlockColor{}
	}
	red, green, blue, alpha := source.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()

	return halfBlockColor{
		red:   uint8(red >> 8),
		green: uint8(green >> 8),
		blue:  uint8(blue >> 8),
		alpha: uint8(alpha >> 8),
	}
}

func resolveURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, errors.New("invalid image URL")
	}
	if parsed.IsAbs() {
		if (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return nil, errors.New("invalid image URL")
		}

		return parsed, nil
	}
	if parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/") {
		return nil, errors.New("invalid image URL")
	}
	base, err := url.Parse(northAPIOrigin)
	if err != nil {
		return nil, err
	}

	return base.ResolveReference(parsed), nil
}

func fit(source image.Image, maximumWidth, maximumHeight int) image.Image {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= maximumWidth && height <= maximumHeight {
		return source
	}
	targetWidth := maximumWidth
	targetHeight := max(1, height*targetWidth/width)
	if targetHeight > maximumHeight {
		targetHeight = maximumHeight
		targetWidth = max(1, width*targetHeight/height)
	}
	target := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for y := range targetHeight {
		sourceY := bounds.Min.Y + y*height/targetHeight
		for x := range targetWidth {
			sourceX := bounds.Min.X + x*width/targetWidth
			target.Set(x, y, source.At(sourceX, sourceY))
		}
	}

	return target
}

func imageKey(rawURL string, columns, rows int) string {
	return fmt.Sprintf("%s\x00%d\x00%d", rawURL, columns, rows)
}

func imageID(key string) int {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))
	id := int(hash.Sum32() & 0x00ffffff)
	if id == 0 {
		return 1
	}

	return id
}
