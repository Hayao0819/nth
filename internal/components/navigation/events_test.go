package navigation

import (
	"testing"

	"github.com/Hayao0819/go-north"
)

func TestOpenCommandsKeepTheirTargets(t *testing.T) {
	t.Parallel()

	post := north.Post{ID: "post-1"}
	if message, ok := OpenPost(post)().(OpenPostMsg); !ok || message.Post.ID != post.ID {
		t.Fatalf("OpenPost() = %#v", message)
	}
	user := north.User{Handle: "alice"}
	if message, ok := OpenUser(user)().(OpenUserMsg); !ok || message.User.Handle != user.Handle {
		t.Fatalf("OpenUser() = %#v", message)
	}
}
