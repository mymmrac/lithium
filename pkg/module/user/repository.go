package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/mymmrac/lithium/pkg/module/db"
	"github.com/mymmrac/lithium/pkg/module/id"
)

type Repository interface {
	Create(ctx context.Context, id id.ID, email, password string) error
	GetByEmail(ctx context.Context, email string) (*Model, bool, error)
}

type repository struct {
	tx db.Transaction
}

func NewRepository(tx db.Transaction) Repository {
	return &repository{
		tx: tx,
	}
}

var ErrAlreadyExists = errors.New("user already exists")

func (r *repository) Create(ctx context.Context, id id.ID, email, password string) error {
	_, err := r.tx.Extract(ctx).NewInsert().Model(&Model{
		ID:       id,
		Email:    email,
		Password: password,
	}).Exec(ctx)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key value violates unique constraint") {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*Model, bool, error) {
	var user Model
	if err := r.tx.Extract(ctx).NewSelect().Model(&user).Where("email = ?", email).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &user, true, nil
}
