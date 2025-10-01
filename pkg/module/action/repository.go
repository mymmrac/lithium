package action

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mymmrac/lithium/pkg/module/db"
	"github.com/mymmrac/lithium/pkg/module/id"
)

type Repository interface {
	Create(ctx context.Context, model *Model) error
	GetByProjectID(ctx context.Context, projectID id.ID) ([]Model, error)
	GetByID(ctx context.Context, id id.ID) (*Model, bool, error)
}

type repository struct {
	tx db.Transaction
}

func NewRepository(tx db.Transaction) Repository {
	return &repository{
		tx: tx,
	}
}

func (r *repository) Create(ctx context.Context, model *Model) error {
	_, err := r.tx.Extract(ctx).NewInsert().Model(model).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetByProjectID(ctx context.Context, projectID id.ID) ([]Model, error) {
	var models []Model
	err := r.tx.Extract(ctx).NewSelect().Model(&models).Where("project_id = ?", projectID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return models, nil
}

func (r *repository) GetByID(ctx context.Context, id id.ID) (*Model, bool, error) {
	var model Model
	if err := r.tx.Extract(ctx).NewSelect().Model(&model).Where("id = ?", id).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &model, true, nil
}
