package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mymmrac/lithium/pkg/handler/auth"
	"github.com/mymmrac/lithium/pkg/handler/static"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/version"
)

func main() {
	cmd := cobra.Command{
		Use:     "lithium",
		Short:   "Lithium is a platform for running WASM modules as lambda functions",
		RunE:    run,
		Version: fmt.Sprintf("%s (%s), built at %s", version.Version(), version.Modified(), version.BuildTime()),
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	v := viper.NewWithOptions()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	v.SetDefault("log-level", "info")
	v.SetDefault("host", "")
	v.SetDefault("port", 4251)

	logger.SetLevel(v.GetString("log-level"))

	logger.Infow(ctx, "starting lithium", "version", version.Version())

	views, err := static.LoadViews()
	if err != nil {
		return fmt.Errorf("load views: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "lithium",
		Views:   views,
	})

	if err = static.RegisterHandlers(app); err != nil {
		return fmt.Errorf("static handler: %w", err)
	}

	apiRouter := app.Group("/api")
	if err = auth.RegisterHandlers(apiRouter); err != nil {
		return fmt.Errorf("auth handler: %w", err)
	}

	go func() {
		defer cancel()

		addr := net.JoinHostPort(v.GetString("host"), v.GetString("port"))
		logger.Infof(ctx, "listening on %s", addr)

		err = app.Listen(addr, fiber.ListenConfig{
			DisableStartupMessage: true,
		})
		if err != nil {
			logger.Errorw(ctx, "run server", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info(ctx, "shutting down...")

	if err = app.ShutdownWithTimeout(time.Second * 5); err != nil {
		logger.Errorw(ctx, "shutting down server", "error", err)
	}

	return nil
}
