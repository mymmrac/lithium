package user

import (
	"github.com/uptrace/bun"

	"github.com/mymmrac/lithium/pkg/module/id"
)

type Model struct {
	bun.BaseModel `bun:"table:user"`

	ID       id.ID  `bun:"id,pk"`
	Email    string `bun:"email"`
	Password string `bun:"password"`
}
