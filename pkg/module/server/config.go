package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"

	"github.com/mymmrac/lithium/pkg/module/di"
)

type Config struct {
	Host string `validate:"omitempty,hostname"`
	Port uint   `validate:"port"`
}

func init() { //nolint:gochecknoinits
	di.Base().MustProvide(func(v *viper.Viper, va *validator.Validate) (Config, error) {
		cfg := Config{
			Host: v.GetString("host"),
			Port: v.GetUint("port"),
		}
		if err := va.Struct(cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	})
}
