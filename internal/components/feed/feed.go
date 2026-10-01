package feed

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/termimage"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type Mode uint8

const (
	Home Mode = iota
	Ranked
	Search
)

func (mode Mode) Label(query string) string {
	switch mode {
	case Ranked:
		return "For you"
	case Search:
		return "Search: " + ui.SafeInline(query)
	default:
		return "Home"
	}
}

type Feed struct {
	reactea.BasicComponent

	api    API
	theme  ui.Theme
	images *termimage.Renderer
	loader Loader
	label  string

	mode  Mode
	query string

	posts      []north.Post
	nextCursor *string
	selected   int
	top        int

	loading      bool
	loadingMore  bool
	fillLoads    int
	refreshReady bool
	seq          uint64
	acting       map[reactionKey]struct{}
	localState   map[string]localPostState
	removed      map[string]struct{}
	err          error
	notice       string
	resp         *north.Response
	respAt       time.Time
}

func New(api API, theme ui.Theme) *Feed {
	return NewWithImages(api, theme, nil)
}

func NewWithImages(api API, theme ui.Theme, images *termimage.Renderer) *Feed {
	return &Feed{
		api:          api,
		theme:        theme,
		images:       images,
		refreshReady: true,
		acting:       make(map[reactionKey]struct{}),
		localState:   make(map[string]localPostState),
		removed:      make(map[string]struct{}),
	}
}

// NewSearch creates an idle feed which starts loading when a query is set.
func NewSearch(api API, theme ui.Theme) *Feed {
	return NewSearchWithImages(api, theme, nil)
}

func NewSearchWithImages(api API, theme ui.Theme, images *termimage.Renderer) *Feed {
	feed := NewWithImages(api, theme, images)
	feed.mode = Search

	return feed
}

func NewSourceWithImages(api API, theme ui.Theme, label string, loader Loader, images *termimage.Renderer) *Feed {
	feed := NewWithImages(api, theme, images)
	feed.label = label
	feed.loader = loader

	return feed
}

func (f *Feed) Init(ctx *reactea.Ctx) tea.Cmd { return f.load(ctx, false) }

var _ reactea.Component = (*Feed)(nil)
