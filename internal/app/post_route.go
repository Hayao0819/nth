package app

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/components/pageheader"
	postfeature "github.com/Hayao0819/nth/internal/features/post"
	"github.com/Hayao0819/nth/internal/ui"
	"github.com/Hayao0819/reactea/v2"
	"github.com/Hayao0819/reactea/v2/state"
)

type postResult struct {
	post north.Post
	resp *north.Response
}

type postRoutePage struct {
	reactea.BasicComponent

	root     *root
	id       string
	resource state.Resource[postResult]
	page     *postfeature.Screen
}

func newPostRoute(root *root, id string) *postRoutePage {
	return &postRoutePage{root: root, id: id}
}

func (p *postRoutePage) Init(ctx *reactea.Ctx) tea.Cmd {
	return p.load(ctx)
}

func (p *postRoutePage) Update(ctx *reactea.Ctx, msg tea.Msg) tea.Cmd {
	if p.resource.Handle(msg) {
		if p.resource.Err() != nil {
			return nil
		}
		result := p.resource.Value()
		p.root.setResponse(result.resp)
		p.root.posts[p.id] = result.post
		p.page = p.root.newPostPage(result.post)

		return p.page.Init(ctx)
	}
	if click, ok := msg.(tea.MouseClickMsg); ok && click.Button == tea.MouseLeft {
		x, y, inside := reactea.Mouse(ctx, msg)
		if inside && pageheader.BackAt(x, y) {
			return pageheader.Back()
		}
	}
	if reactea.Key(msg, "esc", "left") {
		return pageheader.Back()
	}
	if reactea.Key(msg, ".") && p.resource.Err() != nil {
		return p.load(ctx)
	}
	if p.page != nil {
		return p.page.Update(ctx, msg)
	}

	return nil
}

func (p *postRoutePage) Render(ctx *reactea.Ctx) string {
	if p.page != nil {
		return p.page.Render(ctx)
	}

	message := "Loading post…"
	if p.resource.Err() != nil {
		message = ui.FriendlyError(p.resource.Err()) + "\n\nPress . to try again"
	}
	header := pageheader.Render(p.root.theme, "Post", p.root.theme.Dim.Render("Esc/← back"), ctx.Width())
	body := lipgloss.Place(ctx.Width(), max(1, ctx.Height()-pageheader.Height), lipgloss.Center, lipgloss.Center, message)

	return ui.Fit(header+"\n"+body, ctx.Width(), ctx.Height())
}

func (p *postRoutePage) load(ctx *reactea.Ctx) tea.Cmd {
	return p.resource.Load(ctx, func(request context.Context) (postResult, error) {
		post, response, err := p.root.api.Post(request, p.id)

		return postResult{post: post, resp: response}, err
	})
}

func (p *postRoutePage) UpdatePost(id, text string, editedAt time.Time) {
	if p.page != nil {
		p.page.UpdatePost(id, text, editedAt)
	}
}

var _ reactea.Component = (*postRoutePage)(nil)
