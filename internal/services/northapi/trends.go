package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
)

type officialTrendAPI interface {
	Trends(context.Context, string) ([]north.Trend, *north.Response, error)
}

type officialTrendDismissAPI interface {
	DismissTrend(context.Context, string) (bool, *north.Response, error)
}

func (c *Hybrid) Trends(ctx context.Context, cursor string) ([]north.Trend, *north.Response, error) {
	if official, ok := c.official.(officialTrendAPI); ok && c.officialSupportsScopedAPI() {
		return official.Trends(ctx, cursor)
	}

	return withWeb(ctx, c, func(client *Client) ([]north.Trend, *north.Response, error) {
		return client.Trends(ctx, cursor)
	})
}

func (c *Hybrid) DismissTrend(ctx context.Context, tag string) (bool, *north.Response, error) {
	api, ok := c.official.(officialTrendDismissAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return false, nil, errors.New("trend dismissal is not available with the configured credentials")
	}

	return api.DismissTrend(ctx, tag)
}
