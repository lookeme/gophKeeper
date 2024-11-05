package db

import (
	"GophKeeper/internal/models"
	"GophKeeper/utils"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SettingsRepository represents a repository for managing user data.
type SettingsRepository struct {
	postgres *Postgres
}

func NewSettingsRepository(postgres *Postgres) *SettingsRepository {
	return &SettingsRepository{
		postgres: postgres,
	}
}

func (s *SettingsRepository) SaveSettings(ctx context.Context, key string, val string) (uuid.UUID, error) {
	var lastInsertID uuid.UUID
	err := s.postgres.connPool.QueryRow(ctx, "INSERT INTO settings(key, value) VALUES($1, $2) RETURNING id", key, val).Scan(&lastInsertID)
	if err != nil {
		return lastInsertID, err
	}
	return lastInsertID, nil
}

func (s *SettingsRepository) FindSettingsByKey(ctx context.Context, key string) (models.SettingsDTO, error) {
	query := `SELECT key, value FROM settings WHERE key = @key ORDER BY created_at DESC LIMIT 1`
	args := pgx.NamedArgs{
		"key": key,
	}
	var data models.SettingsDTO
	row, err := s.postgres.connPool.Query(ctx, query, args)
	if err != nil {
		return data, err
	}
	data, err = pgx.CollectOneRow(row, pgx.RowToStructByPos[models.SettingsDTO])
	if err != nil {
		return data, err
	}
	return data, nil
}

func (s *SettingsRepository) SaveKeys(ctx context.Context, kek, dek string) error {
	tx, err := s.postgres.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	var version int
	err = tx.QueryRow(ctx, "SELECT version FROM settings ORDER BY created_at DESC LIMIT 1").Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		version = 1
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback(ctx)
	} else if err == nil {
		version += 1
	}
	rows := [][]interface{}{
		{utils.SettingKeyKek, kek, version},
		{utils.SettingKeyDek, dek, version},
	}
	_, txErr := tx.CopyFrom(ctx,
		pgx.Identifier{"settings"},
		[]string{"key", "value", "version"},
		pgx.CopyFromRows(rows),
	)
	if txErr != nil {
		_ = tx.Rollback(ctx)
		return txErr
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil

}
