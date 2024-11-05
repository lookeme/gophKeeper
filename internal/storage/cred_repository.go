package db

import (
	"GophKeeper/internal/models"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DataType int

const (
	CreditCard DataType = iota
	Credentials
)

// CredRepository represents a repository for managing user data.
type CredRepository struct {
	postgres *Postgres
}

func NewCredRepository(postgres *Postgres) *CredRepository {
	return &CredRepository{
		postgres: postgres,
	}
}

func (u *CredRepository) SaveUserCreds(ctx context.Context, credName string, userID string, data string, dataType DataType) (uuid.UUID, error) {
	var lastInsertID uuid.UUID
	err := u.postgres.connPool.QueryRow(
		ctx, "INSERT INTO userscredinfo(user_id, name, data, type) VALUES($1, $2, $3, $4) RETURNING id", userID, credName, data, dataType).Scan(&lastInsertID)
	if err != nil {
		return lastInsertID, err
	}
	return lastInsertID, nil
}

// GetLastUserCreds retrieves the most recent set of user credentials for the given user ID from the database.
func (u *CredRepository) GetLastUserCreds(ctx context.Context, userID string, credName string) (models.UserCredentials, error) {
	query := `SELECT name, user_id, data, type, version, created_at FROM userscredinfo WHERE user_id = @user_id AND name = @name ORDER BY created_at DESC LIMIT 1;`
	args := pgx.NamedArgs{
		"user_id": userID,
		"name":    credName,
	}
	var data models.UserCredentials
	row, err := u.postgres.connPool.Query(ctx, query, args)
	if err != nil {
		return data, err
	}
	data, err = pgx.CollectOneRow(row, pgx.RowToStructByPos[models.UserCredentials])
	if err != nil {
		return data, err
	}
	return data, nil
}

// FindAll retrieves all user credentials from the database and returns them as a slice of UserCredentials.
func (u *CredRepository) FindAll(ctx context.Context, userID string) ([]models.UserCredentials, error) {
	query := `SELECT name, data, type, version, created_at FROM userscredinfo WHERE user_id = $1 ORDER BY version DESC;`
	rows, err := u.postgres.connPool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []models.UserCredentials

	for rows.Next() {
		var cred models.UserCredentials
		if err := rows.Scan(&cred.Name, &cred.Data, &cred.DataType, &cred.Version, &cred.CreatedAt); err != nil {
			return nil, err
		}
		credentials = append(credentials, cred)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return credentials, nil
}
