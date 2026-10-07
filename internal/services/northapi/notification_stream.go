package northapi

import (
	"context"
	"errors"

	"github.com/Hayao0819/go-north"
)

func (c *Hybrid) StreamNotifications(ctx context.Context) (*north.NotificationStream, *north.Response, error) {
	official, ok := c.official.(officialNotificationStreamAPI)
	if !ok || !c.officialSupportsScopedAPI() {
		return nil, nil, errors.New("notification stream is not available with the configured credentials")
	}

	return official.StreamNotifications(ctx)
}
