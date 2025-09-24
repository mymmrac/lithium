package action

import (
	"time"

	"github.com/uptrace/bun"

	"github.com/mymmrac/lithium/pkg/module/id"
)

type Model struct {
	bun.BaseModel `bun:"table:action"`

	ID         id.ID     `bun:"id,pk"`
	ProjectID  id.ID     `bun:"project_id"`
	Name       string    `bun:"name"`
	URL        string    `bun:"url"`
	ModulePath string    `bun:"module_path"`
	CreatedAt  time.Time `bun:"created_at"`
	UpdatedAt  time.Time `bun:"updated_at"`
}
