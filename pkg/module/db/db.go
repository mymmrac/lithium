package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"

	"github.com/mymmrac/lithium/pkg/module/di"
	"github.com/mymmrac/lithium/pkg/module/logger"
)

func init() { //nolint:gochecknoinits
	di.Base().
		MustProvide(func(ctx context.Context, v *viper.Viper) (*bun.DB, error) {
			db := bun.NewDB(sql.OpenDB(pgdriver.NewConnector(
				pgdriver.WithDSN(v.GetString("postgres-connection-string")),
			)), pgdialect.New())

			if err := db.PingContext(ctx); err != nil {
				return nil, fmt.Errorf("ping db: %w", err)
			}

			return db, nil
		}).
		MustProvide(func(db *bun.DB) Transaction { return &transaction{db} })
}

func RunMigrations(ctx context.Context, db *bun.DB) error {
	migrations := migrate.NewMigrations()

	if err := migrations.Discover(os.DirFS("./migrations")); err != nil {
		return fmt.Errorf("discover migrations: %w", err)
	}

	migrator := migrate.NewMigrator(db, migrations,
		migrate.WithTableName("migrations"),
		migrate.WithLocksTableName("migrations_locks"),
	)

	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("migrator init: %w", err)
	}

	if err := migrator.Lock(ctx); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	defer func() {
		if err := migrator.Unlock(ctx); err != nil {
			logger.Errorw(ctx, "unlock migrations", "error", err)
		}
	}()

	if _, err := migrator.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
