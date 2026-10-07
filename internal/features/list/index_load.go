package list

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/Hayao0819/go-north"
	listdomain "github.com/Hayao0819/nth/internal/domain"
)

type indexLoadedMsg struct {
	target     *Index
	collection listdomain.ListCollection
	response   *north.Response
	err        error
}

func (m indexLoadedMsg) Response() *north.Response { return m.response }

func (s *Index) load(ctx context.Context) tea.Cmd {
	if s.api == nil || s.loading {
		return nil
	}
	s.loading = true
	s.err = nil

	return func() tea.Msg {
		collection, response, err := s.api.Lists(ctx, "")

		return indexLoadedMsg{target: s, collection: collection, response: response, err: err}
	}
}

func flatten(collection listdomain.ListCollection) []entry {
	seen := make(map[string]struct{})
	result := make([]entry, 0, len(collection.Pinned)+len(collection.Items)+len(collection.Followed))
	add := func(section string, lists []listdomain.ListItem) {
		for _, item := range lists {
			if item.ID == "" {
				continue
			}
			if _, exists := seen[item.ID]; exists {
				continue
			}
			seen[item.ID] = struct{}{}
			result = append(result, entry{section: section, item: item})
		}
	}
	add("Pinned", collection.Pinned)
	add("Your lists", collection.Items)
	add("Following", collection.Followed)

	return result
}
