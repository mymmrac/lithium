package action

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"

	"github.com/mymmrac/lithium/pkg/module/di"
)

type Config struct {
	ModuleBucket string `validate:"required"`
}

func init() { //nolint:gochecknoinits
	di.Base().MustProvide(func(v *viper.Viper, va *validator.Validate) (Config, error) {
		cfg := Config{
			ModuleBucket: v.GetString("module-bucket"),
		}
		if err := va.Struct(cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	})
}
