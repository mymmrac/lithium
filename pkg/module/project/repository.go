package project

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mymmrac/lithium/pkg/module/db"
	"github.com/mymmrac/lithium/pkg/module/id"
)

type Repository interface {
	Create(ctx context.Context, model *Model) error
	UpdateName(ctx context.Context, id id.ID, name string) error
	GetByID(ctx context.Context, id id.ID) (*Model, bool, error)
	GetByOwnerID(ctx context.Context, ownerID id.ID) ([]Model, error)
	GetBySubDomain(ctx context.Context, subDomain string) (*Model, bool, error)
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

func (r *repository) UpdateName(ctx context.Context, id id.ID, name string) error {
	_, err := r.tx.Extract(ctx).NewUpdate().
		Model(&Model{}).
		Set("name = ?", name).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id id.ID) (*Model, bool, error) {
	var model Model
	err := r.tx.Extract(ctx).NewSelect().
		Model(&model).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &model, true, nil
}

func (r *repository) GetByOwnerID(ctx context.Context, ownerID id.ID) ([]Model, error) {
	var models []Model
	err := r.tx.Extract(ctx).NewSelect().
		Model(&models).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return models, nil
}

func (r *repository) GetBySubDomain(ctx context.Context, subDomain string) (*Model, bool, error) {
	var model Model
	err := r.tx.Extract(ctx).NewSelect().
		Model(&model).
		Where("sub_domain = ?", subDomain).
		Scan(ctx)
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
