package northapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/nth/internal/domain"
)

type profileOfficialStub struct {
	OfficialAPI

	uploads       []string
	deleted       []string
	profile       north.UpdateProfileRequest
	profileErr    error
	uploadPurpose north.MediaPurpose
}

func (s *profileOfficialStub) UpdateProfile(_ context.Context, request north.UpdateProfileRequest) (north.User, *north.Response, error) {
	s.profile = request

	return north.User{ID: "me", Name: "Updated"}, nil, s.profileErr
}

func (s *profileOfficialStub) UploadMediaFile(_ context.Context, path string, options ...north.MediaUploadOption) (north.Media, *north.Response, error) {
	header := make(http.Header)
	for _, option := range options {
		option(header)
	}
	s.uploadPurpose = north.MediaPurpose(header.Get("X-North-Media-Purpose"))
	s.uploads = append(s.uploads, path)

	return north.Media{ID: fmt.Sprintf("media-%d", len(s.uploads))}, nil, nil
}

func (s *profileOfficialStub) DeleteMedia(_ context.Context, id string) (bool, *north.Response, error) {
	s.deleted = append(s.deleted, id)

	return true, nil, nil
}

func TestHybridUpdatesProfileImages(t *testing.T) {
	t.Parallel()

	official := &profileOfficialStub{}
	client := NewHybrid(official, nil, "", nil)
	if !client.SupportsProfileMediaUpdate() {
		t.Fatal("profile media update was not detected")
	}
	user, _, err := client.UpdateProfile(context.Background(), domain.ProfileUpdate{
		Name:       "Updated",
		AvatarPath: "/tmp/avatar.png",
		HeaderPath: "/tmp/header.jpg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "me" {
		t.Errorf("user = %#v", user)
	}
	if !slices.Equal(official.uploads, []string{"/tmp/avatar.png", "/tmp/header.jpg"}) {
		t.Errorf("uploads = %#v", official.uploads)
	}
	if official.uploadPurpose != north.MediaForProfile {
		t.Errorf("media purpose = %q", official.uploadPurpose)
	}
	body, err := json.Marshal(official.profile)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"name":"Updated","bio":null,"location":null,"website":null,"avatarMediaId":"media-1","headerMediaId":"media-2"}`
	if string(body) != want {
		t.Errorf("profile request = %s, want %s", body, want)
	}
	if len(official.deleted) != 0 {
		t.Errorf("successful uploads deleted = %#v", official.deleted)
	}
}

func TestHybridDeletesProfileUploadsWhenUpdateFails(t *testing.T) {
	t.Parallel()

	official := &profileOfficialStub{profileErr: errors.New("update failed")}
	client := NewHybrid(official, nil, "", nil)
	_, _, err := client.UpdateProfile(context.Background(), domain.ProfileUpdate{
		Name:       "Updated",
		AvatarPath: "/tmp/avatar.png",
		HeaderPath: "/tmp/header.jpg",
	})
	if !errors.Is(err, official.profileErr) {
		t.Fatalf("UpdateProfile error = %v", err)
	}
	if !slices.Equal(official.deleted, []string{"media-1", "media-2"}) {
		t.Errorf("deleted = %#v", official.deleted)
	}
}
