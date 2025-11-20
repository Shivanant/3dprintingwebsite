package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func Migrate(ctx context.Context, db *gorm.DB) error {
	if err := enableExtensions(ctx, db); err != nil {
		return err
	}
	return db.WithContext(ctx).AutoMigrate(AllModels()...)
}

func enableExtensions(ctx context.Context, db *gorm.DB) error {
	raw := `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
`
	if err := db.WithContext(ctx).Exec(raw).Error; err != nil {
		return fmt.Errorf("enable extensions: %w", err)
	}
	return nil
}
