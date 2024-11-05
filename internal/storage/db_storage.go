package db

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (

	// pgInstance is a pointer to a Postgres struct variable used for storing the initialized instance of the Postgres database connection pool and logger.
	pgInstance *Postgres

	// pgOnce is a sync.Once variable used for lazy initialization of the pgInstance variable in the New function.
	pgOnce sync.Once
)

// Postgres is a type representing a connection pool to a PostgreSQL database.
// It contains a *pgxpool.Pool object for managing connections and a *logger.Logger object for logging.
type Postgres struct {
	connPool *pgxpool.Pool
	log      *zap.Logger
}

// Storage Implementation omitted for brevity
type Storage struct {
	UserRepository     *UserRepository
	SettingsRepository *SettingsRepository
	CredRepository     *CredRepository
}

// NewStorage creates a new instance of Storage by accepting an implementation of UserRepository and ShortenRepository.
func NewStorage(userRepo *UserRepository, settingsRepo *SettingsRepository, credRepo *CredRepository) *Storage {
	return &Storage{
		UserRepository:     userRepo,
		SettingsRepository: settingsRepo,
		CredRepository:     credRepo,
	}
}

// Close closes the PostgreSQL connection pool by calling the Close method of the connPool.
// It does not return any error.
func (pg *Postgres) Close() error {
	pg.connPool.Close()
	return nil
}

// Ping pings the Postgres database by calling the Ping method of the connPool.
func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.connPool.Ping(ctx)
}

// New creates a new instance of Postgres by accepting a context, logger, and storage configuration.
func New(ctx context.Context, log *zap.Logger, connStr string) (*Postgres, error) {
	log.Info("creating pool of conn to db...", zap.String("connString", connStr))
	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, connStr)
		if err != nil {
			log.Error(err.Error())
		}
		pgInstance = &Postgres{db, log}
	})
	err := StartMigration(pgInstance.connPool)
	if err != nil {
		return nil, err
	}
	return pgInstance, nil
}
