package server

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/mymmrac/lithium/pkg/module/di"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/runner"
)

type Server runner.Service

func init() { //nolint:gochecknoinits
	di.Base().
		MustProvide(func(ctx context.Context, app *fiber.App, server Server, runner runner.Runner) fiber.Router {
			runner.Add(ctx, server)
			return app
		}).
		MustProvide(NewFiberServer)
}

type FiberServer struct {
	cfg Config
	app *fiber.App
}

func NewFiberServer(cfg Config, app *fiber.App) Server {
	return &FiberServer{
		cfg: cfg,
		app: app,
	}
}

func (f *FiberServer) Run(ctx context.Context) error {
	addr := net.JoinHostPort(f.cfg.Host, strconv.FormatUint(uint64(f.cfg.Port), 10))
	logger.Infof(ctx, "listening on %s", addr)

	err := f.app.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: true,
	})
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

func (f *FiberServer) Stop() {
	if err := f.app.ShutdownWithTimeout(time.Second * 5); err != nil {
		logger.Errorw(context.Background(), "shutting down server", "error", err)
	}
}
