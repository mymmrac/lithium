package storage

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"

	"github.com/mymmrac/lithium/pkg/module/di"
)

type Config struct {
	Endpoint string `validate:"hostname"`
	Secure   bool   `validate:"-"`
	ID       string `validate:"required"`
	Secret   string `validate:"required"`
}

func init() { //nolint:gochecknoinits
	di.Base().MustProvide(func(v *viper.Viper, va *validator.Validate) (Config, error) {
		cfg := Config{
			Endpoint: v.GetString("minio-endpoint"),
			Secure:   v.GetBool("minio-secure"),
			ID:       v.GetString("minio-id"),
			Secret:   v.GetString("minio-secret"),
		}
		if err := va.Struct(cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	})
}
