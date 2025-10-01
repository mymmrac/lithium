package validator

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"

	"github.com/mymmrac/lithium/pkg/module/di"
)

func init() { //nolint:gochecknoinits
	di.Base().MustProvide(NewValidate)
}

func NewValidate() (*validator.Validate, error) {
	v := validator.New(validator.WithRequiredStructEnabled())

	alphaNumTextRegexp := regexp.MustCompile("^[a-zA-Z0-9_ -]+$")
	err := v.RegisterValidation("alphanum_text", func(fl validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}
		return alphaNumTextRegexp.MatchString(value)
	})
	if err != nil {
		return nil, fmt.Errorf("register alphanum_text validator: %w", err)
	}

	return v, nil
}
