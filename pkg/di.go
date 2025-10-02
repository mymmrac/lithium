package pkg

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rathil/rdi"
	"github.com/spf13/viper"

	"github.com/mymmrac/lithium/pkg/handler/invoker"
	"github.com/mymmrac/lithium/pkg/handler/static"
	"github.com/mymmrac/lithium/pkg/module/action"
	"github.com/mymmrac/lithium/pkg/module/auth"
	"github.com/mymmrac/lithium/pkg/module/di"
	"github.com/mymmrac/lithium/pkg/module/project"
	"github.com/mymmrac/lithium/pkg/module/storage"
	"github.com/mymmrac/lithium/pkg/module/user"
	"github.com/mymmrac/lithium/pkg/module/version"
)

func DI(ctx context.Context, v *viper.Viper) rdi.DI {
	return di.New(ctx, v).
		MustProvide(func(v *validator.Validate, views static.Views, auth auth.Auth, invoker invoker.Invoker) *fiber.App {
			app := fiber.New(fiber.Config{
				AppName:         version.Name(),
				Views:           views,
				StructValidator: &FiberValidatorAdapter{v: v},
				BodyLimit:       64 * 1024 * 1024,
			})
			app.Use(invoker.Middleware)
			app.Use(auth.Middleware)
			return app
		}).
		MustProvide(static.LoadViews).
		MustProvide(auth.NewAuth).
		MustProvide(invoker.NewInvoker).
		MustProvide(storage.NewStorage).
		MustProvide(user.NewRepository).
		MustProvide(project.NewRepository).
		MustProvide(action.NewRepository)
}

type FiberValidatorAdapter struct {
	v *validator.Validate
}

func (v *FiberValidatorAdapter) Validate(value any) error {
	return v.v.Struct(value)
}
