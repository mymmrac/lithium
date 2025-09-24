package project

import (
	"time"

	"github.com/uptrace/bun"

	"github.com/mymmrac/lithium/pkg/module/id"
)

type Model struct {
	bun.BaseModel `bun:"table:project"`

	ID        id.ID     `bun:"id,pk"`
	OwnerID   id.ID     `bun:"owner_id"`
	Name      string    `bun:"name"`
	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
}
