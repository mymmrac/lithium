package di

import (
	"context"

	"github.com/rathil/rdi"
	"github.com/rathil/rdi/standard"
	"github.com/spf13/viper"
)

//nolint:gochecknoglobals
var base = standard.NewWithParent(nil)

// Base returns the default global DI container.
func Base() rdi.DI {
	return base
}

// New creates a new DI container with the Base container as its parent.
func New(ctx context.Context, v *viper.Viper) rdi.DI {
	ctx, cancel := context.WithCancel(ctx)
	standard.SetTraceLevel(standard.TraceFilePath)
	return standard.NewWithParent(base).
		MustProvide(func() context.Context { return ctx }).
		MustProvide(func() context.CancelFunc { return cancel }).
		MustProvide(v)
}
