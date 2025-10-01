package action

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/mymmrac/lithium/pkg/module/db"
	"github.com/mymmrac/lithium/pkg/module/id"
)

type Repository interface {
	Create(ctx context.Context, model *Model) error
	UpdateInfo(ctx context.Context, id id.ID, name, path string, methods []string) error
	GetByID(ctx context.Context, id id.ID) (*Model, bool, error)
	GetByProjectID(ctx context.Context, projectID id.ID) ([]Model, error)
	DeleteByID(ctx context.Context, id id.ID) error
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

func (r *repository) UpdateInfo(ctx context.Context, id id.ID, name, path string, methods []string) error {
	_, err := r.tx.Extract(ctx).NewUpdate().
		Model(&Model{}).
		Set("name = ?", name).
		Set("path = ?", path).
		Set("methods = ?", pgdialect.Array(methods)).
		Where("id = ?", id).
		Exec(ctx)
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
	err := r.tx.Extract(ctx).NewSelect().Model(&model).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &model, true, nil
}

func (r *repository) DeleteByID(ctx context.Context, id id.ID) error {
	var model Model
	_, err := r.tx.Extract(ctx).NewDelete().
		Model(&model).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
