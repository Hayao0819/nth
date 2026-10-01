package post

import (
	"strings"

	"github.com/Hayao0819/go-north"
)

// LinkedUsers returns the accounts whose names are interactive in a post.
func LinkedUsers(post north.Post, related ...north.Post) []north.User {
	users := make([]north.User, 0, 4+len(related))
	seen := make(map[string]struct{})
	add := func(user north.User) {
		user.Handle = strings.TrimPrefix(strings.TrimSpace(user.Handle), "@")
		if user.Handle == "" {
			return
		}
		key := strings.ToLower(user.Handle)
		if user.ID != "" {
			key = "id:" + user.ID
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		users = append(users, user)
	}

	target := post.DisplayPost()
	if target != nil {
		add(target.Author)
	}
	if post.RepostOf != nil {
		add(post.Author)
	}
	if target != nil && target.InReplyToHandle != nil {
		add(north.User{Handle: *target.InReplyToHandle})
	}
	if target != nil && target.Quoted != nil {
		quoted := target.Quoted.DisplayPost()
		if quoted != nil {
			add(quoted.Author)
		}
	}
	for _, item := range related {
		if itemTarget := item.DisplayPost(); itemTarget != nil {
			add(itemTarget.Author)
		}
	}

	return users
}
