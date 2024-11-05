package db

import (
	"GophKeeper/internal/models"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// UserRepository represents a repository for managing user data.
type UserRepository struct {
	postgres *Postgres
}

// NewUserRepository creates a new instance of UserRepository with the given Postgres instance.
// It takes a reference to a Postgres instance and returns a pointer to a UserRepository.
func NewUserRepository(postgres *Postgres) *UserRepository {
	return &UserRepository{
		postgres: postgres,
	}
}

// SaveUser saves a new user with the given name and password to the database.
// It takes two string arguments: name and pass.
// It returns the last inserted ID as an integer and an error if the insertion fails.
// The function first prepares an INSERT statement to insert the user into the "users" table, returning the ID of the newly inserted row.
// The values of the name and pass arguments are used as the parameters for the INSERT statement.
// If the insertion is successful, the last inserted ID is scanned into the lastInsertID variable.
// If there is an error during the insertion, the function returns the lastInsertID and the error.
func (u *UserRepository) SaveUser(ctx context.Context, username string, password string, email string) (uuid.UUID, error) {
	var lastInsertID uuid.UUID
	err := u.postgres.connPool.QueryRow(
		ctx, "INSERT INTO users(username, password, email) VALUES($1, $2, $3) RETURNING id",
		username, password, email).Scan(&lastInsertID)
	if err != nil {
		return lastInsertID, err
	}
	return lastInsertID, nil
}

// FindByName finds a user in the database by their userName.
// It takes an integer argument userID, which represents the ID of the user to find.
// The function returns a models.User object and an error.
// It currently returns an empty models.User object and nil error.
// To retrieve the user from the database, the function executes a query with the given userID and scans the result into a models.User object.
// If no user is found with the given userID, the function returns an empty models.User object.
// In case of any error during the query execution, the function returns the empty models.User object and the error.
func (u *UserRepository) FindByName(ctx context.Context, userName string) (models.UserDTO, error) {
	query := `SELECT id, username, email, password, role, created_at, updated_at FROM users WHERE username = @username`
	args := pgx.NamedArgs{
		"username": userName,
	}
	var data models.UserDTO
	row, err := u.postgres.connPool.Query(ctx, query, args)
	if err != nil {
		return data, err
	}
	data, err = pgx.CollectOneRow(row, pgx.RowToStructByPos[models.UserDTO])
	if err != nil {
		return data, err
	}
	return data, nil
}

func (u *UserRepository) FindByEmail(ctx context.Context, email string) (models.UserDTO, error) {
	query := `SELECT id, username, email, role, created_at FROM users WHERE email = @email`
	args := pgx.NamedArgs{
		"email": email,
	}
	var data models.UserDTO
	row, err := u.postgres.connPool.Query(ctx, query, args)
	if err != nil {
		return data, err
	}
	data, err = pgx.CollectOneRow(row, pgx.RowToStructByPos[models.UserDTO])
	if err != nil {
		return data, err
	}
	return data, nil
}
