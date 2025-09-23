package runner

import (
	"context"
	"errors"
)

type contextWatcher struct {
	cancel context.CancelFunc
}

func (c *contextWatcher) Run(ctx context.Context) error {
	<-ctx.Done()
	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (c *contextWatcher) Stop() {
	c.cancel()
}
