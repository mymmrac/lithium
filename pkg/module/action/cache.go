package action

import (
	extism "github.com/extism/go-sdk"

	"github.com/mymmrac/lithium/pkg/module/cache"
	"github.com/mymmrac/lithium/pkg/module/id"
)

type Module struct {
	CompiledPlugin       *extism.CompiledPlugin
	PluginInstanceConfig extism.PluginInstanceConfig
}

type Cache cache.Cache[id.ID, Module]

func NewCache() Cache {
	return cache.NewInMemory[id.ID, Module]()
}
