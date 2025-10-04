package action

import (
	"time"

	"github.com/uptrace/bun"

	"github.com/mymmrac/lithium/pkg/module/id"
)

type Model struct {
	bun.BaseModel `bun:"table:action"`

	ID         id.ID        `bun:"id,pk"`
	ProjectID  id.ID        `bun:"project_id"`
	Name       string       `bun:"name"`
	Path       string       `bun:"path"`
	Methods    []string     `bun:"methods,array"`
	Order      int          `bun:"order"`
	ModulePath string       `bun:"module_path"`
	Config     ModuleConfig `bun:"config,type:jsonb"`
	CreatedAt  time.Time    `bun:"created_at"`
	UpdatedAt  time.Time    `bun:"updated_at"`
}

type ModuleConfig struct {
	Envs    map[string]string `json:"envs,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Network bool              `json:"network,omitempty"`
}
