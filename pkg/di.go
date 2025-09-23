package pkg

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/rathil/rdi"
	"github.com/spf13/viper"

	"github.com/mymmrac/lithium/pkg/handler/static"
	"github.com/mymmrac/lithium/pkg/module/di"
	"github.com/mymmrac/lithium/pkg/module/version"

	_ "github.com/mymmrac/lithium/pkg/module/runner"
	_ "github.com/mymmrac/lithium/pkg/module/server"
	_ "github.com/mymmrac/lithium/pkg/module/validator"
)

func DI(ctx context.Context, v *viper.Viper) rdi.DI {
	return di.New(ctx, v).
		MustProvide(func(views static.Views) *fiber.App {
			return fiber.New(fiber.Config{
				AppName: version.Name(),
				Views:   views,
			})
		}).
		MustProvide(static.LoadViews)
}
