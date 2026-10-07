package northapi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Hayao0819/go-north"
	"github.com/Hayao0819/go-north/unofficial"
	"github.com/Hayao0819/nth/internal/domain"
)

type officialUserActivityAPI interface {
	LikedPosts(context.Context, string, string) (north.PostPage, *north.Response, error)
}

type officialUserConnectionsAPI interface {
	Followers(context.Context, string, string) (north.UserPage, *north.Response, error)
	Following(context.Context, string, string) (north.UserPage, *north.Response, error)
}

type officialProfileAPI interface {
	UpdateProfile(context.Context, north.UpdateProfileRequest) (north.User, *north.Response, error)
}

type officialProfileMediaAPI interface {
	UploadMediaFile(context.Context, string, ...north.MediaUploadOption) (north.Media, *north.Response, error)
	DeleteMedia(context.Context, string) (bool, *north.Response, error)
}

func (c *Hybrid) SupportsUserActivity() bool {
	_, supported := c.official.(officialUserActivityAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsUserConnections() bool {
	_, supported := c.official.(officialUserConnectionsAPI)

	return supported && c.officialSupportsScopedAPI()
}

func (c *Hybrid) SupportsProfileUpdate() bool {
	_, official := c.official.(officialProfileAPI)

	return (official && c.officialSupportsScopedAPI()) || c.webConfigured()
}

func (c *Hybrid) SupportsProfileMediaUpdate() bool {
	_, profile := c.official.(officialProfileAPI)
	_, media := c.official.(officialProfileMediaAPI)

	return profile && media && c.officialSupportsScopedAPI()
}

func (c *Hybrid) LikedPosts(ctx context.Context, handle, cursor string) (north.PostPage, *north.Response, error) {
	api, ok := c.official.(officialUserActivityAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.PostPage{}, nil, errors.New("liked posts are not available with the configured credentials")
	}

	return api.LikedPosts(ctx, handle, cursor)
}

func (c *Hybrid) Followers(ctx context.Context, handle, cursor string) (north.UserPage, *north.Response, error) {
	api, ok := c.official.(officialUserConnectionsAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.UserPage{}, nil, errors.New("followers are not available with the configured credentials")
	}

	return api.Followers(ctx, handle, cursor)
}

func (c *Hybrid) Following(ctx context.Context, handle, cursor string) (north.UserPage, *north.Response, error) {
	api, ok := c.official.(officialUserConnectionsAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return north.UserPage{}, nil, errors.New("following accounts are not available with the configured credentials")
	}

	return api.Following(ctx, handle, cursor)
}

func (c *Hybrid) UpdateProfile(ctx context.Context, update domain.ProfileUpdate) (north.User, *north.Response, error) {
	if api, ok := c.official.(officialProfileAPI); ok && c.officialSupportsScopedAPI() {
		name := update.Name
		request := north.UpdateProfileRequest{
			Name:     &name,
			Bio:      nullableString(update.Bio),
			Location: nullableString(update.Location),
			Website:  nullableString(update.Website),
		}
		var uploaded []string
		for _, image := range []struct {
			label string
			path  string
			set   func(string)
		}{
			{label: "avatar", path: update.AvatarPath, set: func(id string) { request.AvatarMediaID = north.NewNullableString(id) }},
			{label: "header", path: update.HeaderPath, set: func(id string) { request.HeaderMediaID = north.NewNullableString(id) }},
		} {
			if strings.TrimSpace(image.path) == "" {
				continue
			}
			media, response, err := c.uploadProfileImage(ctx, image.path)
			if err != nil {
				c.deleteProfileUploads(ctx, uploaded)

				return north.User{}, response, fmt.Errorf("upload %s: %w", image.label, err)
			}
			image.set(media.ID)
			uploaded = append(uploaded, media.ID)
		}

		user, response, err := api.UpdateProfile(ctx, request)
		if err != nil {
			c.deleteProfileUploads(ctx, uploaded)
		}

		return user, response, err
	}

	return withWeb(ctx, c, func(client *Client) (north.User, *north.Response, error) {
		return client.UpdateProfile(ctx, update)
	})
}

func (c *Client) UpdateProfile(ctx context.Context, update domain.ProfileUpdate) (north.User, *north.Response, error) {
	if update.AvatarPath != "" || update.HeaderPath != "" {
		return north.User{}, nil, errors.New("profile image updates require API authentication")
	}
	name := update.Name
	response, err := c.web.UpdateProfile(ctx, unofficial.UpdateProfileRequest{
		Name:     &name,
		Bio:      webNullableString(update.Bio),
		Location: webNullableString(update.Location),
		Website:  webNullableString(update.Website),
	})
	if err != nil {
		return north.User{}, publicResponse(response), err
	}
	user, fetched, err := c.Me(ctx)
	if fetched != nil {
		return user, fetched, err
	}

	return user, publicResponse(response), err
}

func (c *Hybrid) uploadProfileImage(ctx context.Context, path string) (north.Media, *north.Response, error) {
	api, ok := c.official.(officialProfileMediaAPI)
	if !ok {
		return north.Media{}, nil, errors.New("profile image updates are not available with the configured credentials")
	}
	media, response, err := api.UploadMediaFile(ctx, path, north.WithMediaPurpose(north.MediaForProfile))
	if err == nil && media.ID == "" {
		err = errors.New("media upload returned an empty ID")
	}

	return media, response, err
}

func (c *Hybrid) deleteProfileUploads(ctx context.Context, ids []string) {
	api, ok := c.official.(officialProfileMediaAPI)
	if !ok {
		return
	}
	ctx = context.WithoutCancel(ctx)
	for _, id := range ids {
		_, _, _ = api.DeleteMedia(ctx, id)
	}
}

func nullableString(value string) *north.NullableString {
	if value == "" {
		return north.NullString()
	}

	return north.NewNullableString(value)
}

func webNullableString(value string) **string {
	var nullable *string
	if value != "" {
		copy := value
		nullable = &copy
	}

	return &nullable
}
