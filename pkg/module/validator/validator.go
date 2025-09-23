package validator

import (
	"github.com/go-playground/validator/v10"

	"github.com/mymmrac/lithium/pkg/module/di"
)

func init() { //nolint:gochecknoinits
	di.Base().MustProvide(NewValidate)
}

func NewValidate() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}
