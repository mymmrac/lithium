package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mymmrac/lithium/pkg"
	"github.com/mymmrac/lithium/pkg/handler/auth"
	"github.com/mymmrac/lithium/pkg/handler/static"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/runner"
	_ "github.com/mymmrac/lithium/pkg/module/server"
	_ "github.com/mymmrac/lithium/pkg/module/validator"
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
		os.Exit(1) //nolint:gocritic
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
	err := pkg.DI(ctx, v).
		Invoke(
			static.RegisterHandlers,
			auth.RegisterHandlers,
			runner.RunAndWait,
		)
	if err != nil {
		return err
	}
	logger.Info(ctx, "shutting down...")

	return nil
}
