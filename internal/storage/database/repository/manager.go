package repository

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

// Manager wraps SQLC queries with connection management
type Manager struct {
	Users      UserRepositoryInterface
	Movies     MovieRepositoryInterface
	TVShows    TVShowRepositoryInterface
	Watchlists WatchlistRepositoryInterface
	rawQueries *database.Queries
	pool       *pgxpool.Pool
}

type Tx struct {
	tx    pgx.Tx
	Repos *ReposTx
}

type ReposTx struct {
	Users      UserRepositoryInterface
	Movies     MovieRepositoryInterface
	TVShows    TVShowRepositoryInterface
	Watchlists WatchlistRepositoryInterface
}

// connectSqlcWithPool connects to the database and returns a SQLC Queries instance with the underlying pool
func connectSqlcWithPool(config *config.Config, ctx context.Context) (*Manager, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %v", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second

	// Connect to the database with a timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctxWithTimeout, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	log.Printf("Successfully connected to database %s on %s:%s",
		config.Database.Name,
		config.Database.Host,
		config.Database.Port,
	)

	return &Manager{
		Users:      NewUserRepository(pool),
		Movies:     NewMovieRepository(pool),
		TVShows:    NewTVShowRepository(pool),
		Watchlists: NewWatchlistRepository(pool),
		rawQueries: database.New(pool),
		pool:       pool,
	}, nil
}

func MustConnectDB(config *config.Config, ctx context.Context) *Manager {
	db, err := connectSqlcWithPool(config, ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

// Close closes the database connection pool
func (m *Manager) Close() {
	if m.pool != nil {
		m.pool.Close()
	}
}

func (m *Manager) RawSql() *database.Queries {
	return m.rawQueries
}

// BeginTx starts a transaction and returns a wrapped Tx containing the repos.
func (m *Manager) BeginTx(ctx context.Context) (*Tx, error) {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Tx{
		tx: tx,
		Repos: &ReposTx{
			Users:      NewUserRepository(tx),
			Movies:     NewMovieRepository(tx),
			TVShows:    NewTVShowRepository(tx),
			Watchlists: NewWatchlistRepository(tx),
		},
	}, nil
}

// Commit commits the transaction and clears the internal transaction pointer.
func (t *Tx) Commit(ctx context.Context) error {
	err := t.tx.Commit(ctx)
	t.tx = nil // Prevent deferred rollback
	return err
}

// Rollback rolls back the transaction if it hasn't been committed.
func (t *Tx) Rollback(ctx context.Context) error {
	if t.tx == nil {
		return nil // Already committed or rolled back.
	}
	err := t.tx.Rollback(ctx)
	t.tx = nil
	return err
}
