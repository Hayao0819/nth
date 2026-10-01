package feed

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

func (f *Feed) Render(ctx *reactea.Ctx) string {
	width, height := ctx.Size()
	if width <= 0 || height <= 0 {
		return ""
	}
	if len(f.posts) == 0 {
		return f.renderEmpty(width, height)
	}

	var cards []string
	used := 0
	for index := f.top; index < len(f.posts); index++ {
		if used >= height {
			break
		}
		card := postcomponent.RenderCardWithImages(f.posts[index], width, index == f.selected, f.theme, time.Now(), f.images)
		cards = append(cards, card)
		used += lipgloss.Height(card)
	}
	if f.loadingMore && used < height {
		cards = append(cards, f.theme.Dim.Render("  Loading older posts…"))
	}

	return ui.Fit(strings.Join(cards, "\n"), width, height)
}

func (f *Feed) renderEmpty(width, height int) string {
	message := "No posts here"
	hint := "Press ↑ to refresh or / to search"
	label := f.mode.Label(f.query)
	if f.label != "" {
		label = f.label
	}
	switch {
	case f.loading:
		message = "Loading " + label + "…"
		hint = ""
	case f.err != nil:
		message = ui.FriendlyError(f.err)
		hint = "Press ↑ to try again"
	case f.mode == Search:
		message = "No posts found"
		hint = "Type to change the search or press Esc to go back"
	}

	content := "\n  " + f.theme.Heading.Render(message)
	if hint != "" {
		content += "\n  " + f.theme.Dim.Render(hint)
	}

	return ui.Fit(content, width, height)
}
