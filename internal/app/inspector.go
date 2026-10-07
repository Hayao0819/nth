package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/nth/internal/components/navigation"
	postcomponent "github.com/Hayao0819/nth/internal/components/post"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
)

type inspector struct {
	reactea.BasicComponent
	root *root
}

type inspectorLayout struct {
	content           string
	searchFirst       int
	searchLast        int
	trends            []trendHit
	postFirst         int
	postLast          int
	authorLeft        int
	authorNameRow     int
	authorNameRight   int
	authorHandleRow   int
	authorHandleRight int
	actionFirst       int
	actionLast        int
}

type trendHit struct {
	first       int
	last        int
	dismissLeft int
	query       string
	tag         string
}

func (l inspectorLayout) authorHit(x, y int) bool {
	return x >= l.authorLeft &&
		(y == l.authorNameRow && x < l.authorNameRight ||
			y == l.authorHandleRow && x < l.authorHandleRight)
}

func (i *inspector) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || click.Button != tea.MouseLeft {
		return nil
	}
	x, y, inside := reactea.Mouse(ctx, msg)
	if !inside {
		return nil
	}
	layout := i.build(ctx.Width(), ctx.Height())
	if y >= layout.searchFirst && y <= layout.searchLast {
		return func() tea.Msg { return navigationMsg{action: navigateSearch} }
	}
	for _, hit := range layout.trends {
		if y >= hit.first && y <= hit.last {
			if x >= hit.dismissLeft && i.root.trendDismiss != nil {
				return i.root.dismissTrend(ctx.Context(), hit.tag)
			}
			return func() tea.Msg {
				return navigationMsg{action: navigateSearch, query: hit.query}
			}
		}
	}
	post := i.root.feed.SelectedPost()
	if post == nil {
		return nil
	}
	if layout.authorHit(x, y) {
		target := post.DisplayPost()
		if target == nil {
			return nil
		}
		user := target.Author

		return navigation.OpenUser(user)
	}
	if y >= layout.postFirst && y <= layout.postLast {
		selected := *post

		return navigation.OpenPost(selected)
	}
	if y < layout.actionFirst || y > layout.actionLast {
		return nil
	}
	action := postcomponent.Reply
	if y == layout.actionFirst {
		if x >= ctx.Width()/2 {
			action = postcomponent.Repost
		}
	} else if x < ctx.Width()/2 {
		action = postcomponent.Like
	} else {
		action = postcomponent.Quote
	}
	selected := *post

	return postcomponent.Request(action, selected)
}

func (i *inspector) Render(ctx *reactea.Ctx) string {
	return i.build(ctx.Width(), ctx.Height()).content
}

func (i *inspector) build(width, height int) inspectorLayout {
	r := i.root
	result := inspectorLayout{
		searchFirst: -1, searchLast: -1,
		postFirst: -1, postLast: -1,
		authorNameRow: -1, authorHandleRow: -1,
		actionFirst: -1, actionLast: -1,
	}
	lines := []string{""}
	search := r.theme.Box.
		Width(max(1, width-4)).
		Padding(0, 1).
		MarginLeft(2).
		Render(r.theme.Section.Render("/  Search north"))
	result.searchFirst = len(lines)
	searchLines := strings.Split(search, "\n")
	lines = append(lines, searchLines...)
	result.searchLast = len(lines) - 1
	lines = append(lines, "", "  "+ui.Sides(
		r.theme.Section.Render("TRENDS FOR YOU"),
		r.theme.Dim.Render("6–0"),
		max(1, width-4),
	))
	if r.trendBusy && len(r.trends) == 0 {
		lines = append(lines, "", "  "+r.theme.Dim.Render("Loading trends…"))
	} else if r.trendErr != nil && len(r.trends) == 0 {
		lines = append(lines, "", "  "+r.theme.Warn.Render("Trends unavailable"), "  "+r.theme.Dim.Render("Press . to retry"))
	} else {
		for index, trend := range r.trends[:min(5, len(r.trends))] {
			query := trend.Tag
			label := trend.Tag
			if trend.IsHashtag && !strings.HasPrefix(query, "#") {
				query = "#" + query
				label = query
			}
			first := len(lines)
			right := fmt.Sprintf("%d posts", trend.Count)
			if r.trendDismiss != nil {
				right += "  ×"
			}
			lines = append(lines,
				"  "+r.theme.Dim.Render(fmt.Sprintf("%s · Trending", [...]string{"6", "7", "8", "9", "0"}[index])),
				"  "+ui.Sides(r.theme.Heading.Render(ui.Clip(ui.SafeInline(label), max(1, width-lipgloss.Width(right)-6))), r.theme.Dim.Render(right), max(1, width-4)),
			)
			result.trends = append(result.trends, trendHit{
				first: first, last: len(lines) - 1, dismissLeft: max(0, width-4), query: query, tag: trend.Tag,
			})
		}
	}

	if len(lines)+7 > height {
		result.content = ui.Fit(strings.Join(lines, "\n"), width, height)

		return result
	}
	lines = append(lines, "", "  "+ui.Sides(
		r.theme.Section.Render("SELECTED POST"),
		r.theme.Dim.Render("u/U · Enter"),
		max(1, width-4),
	))

	if post := r.feed.SelectedPost(); post != nil && post.DisplayPost() != nil {
		target := post.DisplayPost()
		cardWidth := max(1, width-8)
		author := r.theme.Name.Render(ui.Clip(ui.SafeInline(target.Author.Name), cardWidth))
		handle := r.theme.Handle.Render(ui.Clip("@"+ui.SafeInline(target.Author.Handle), cardWidth))
		body := ui.WrappedLines(target.Text, cardWidth)
		if postcomponent.Concealed(target) {
			body = ui.WrappedLines(postcomponent.ConcealmentLabel(target), cardWidth)
		}
		if len(body) == 0 && len(target.Media) > 0 {
			body = []string{"[media]"}
		}
		cardLines := []string{author, handle}
		cardLines = append(cardLines, body[:min(3, len(body))]...)
		cardLines = append(cardLines, r.theme.Dim.Render(ui.Clip(fmt.Sprintf("↩ %d   ↻ %d   ♡ %d", target.ReplyCount, target.RepostCount, target.LikeCount), cardWidth)))
		card := r.theme.Box.
			Width(max(1, width-4)).
			Padding(0, 1).
			MarginLeft(2).
			Render(strings.Join(cardLines, "\n"))
		lines = append(lines, "")
		result.postFirst = len(lines)
		result.authorLeft = 4
		result.authorNameRow = result.postFirst + 1
		result.authorNameRight = result.authorLeft + lipgloss.Width(author)
		result.authorHandleRow = result.postFirst + 2
		result.authorHandleRight = result.authorLeft + lipgloss.Width(handle)
		lines = append(lines, strings.Split(card, "\n")...)
		result.postLast = len(lines) - 1
		if len(lines)+4 > height {
			result.content = ui.Fit(strings.Join(lines, "\n"), width, height)

			return result
		}
		lines = append(lines, "", "  "+r.theme.Section.Render("ACTIONS"))

		if postcomponent.CanInteract(post) {
			repost := r.theme.Heading.Render("t  ↻ Repost")
			like := r.theme.Heading.Render("l  ♡ Like")
			if target.Reposted {
				repost = r.theme.Active.Render("t  ↻ Undo")
			}
			if target.Liked {
				like = r.theme.Active.Render("l  ♥ Unlike")
			}
			result.actionFirst = len(lines)
			lines = append(lines,
				ui.Columns([]string{r.theme.Heading.Render("r  ↩ Reply"), repost}, width),
				ui.Columns([]string{like, r.theme.Heading.Render("Q  ❝ Quote")}, width),
			)
			result.actionLast = len(lines) - 1
		} else {
			lines = append(lines, "  "+r.theme.Dim.Render("Unavailable for this post"))
		}
	} else {
		lines = append(lines, "", "  "+r.theme.Dim.Render("No post selected"))
	}

	result.content = ui.Fit(strings.Join(lines, "\n"), width, height)

	return result
}

var _ reactea.Component = (*inspector)(nil)
