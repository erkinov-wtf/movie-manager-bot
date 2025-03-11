package database

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"time"
)

// DbConn wraps SQLC queries with connection management
type DbConn struct {
	*Queries
	pool *pgxpool.Pool
}

// Close closes the database connection pool
func (db *DbConn) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func MustLoadDb(config *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	log.Print("DB connected successfully")

	err = db.AutoMigrate(&models.Movie{}, &models.TVShows{}, &models.User{}, &models.Watchlist{})
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate models: %v", err))
	}

	log.Print("Models migrated successfully")

	return db
}

// ConnectSqlcWithPool connects to the database and returns a SQLC Queries instance with the underlying pool
func ConnectSqlcWithPool(config *config.Config, ctx context.Context) (*DbConn, error) {
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

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

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

	return &DbConn{
		Queries: New(pool),
		pool:    pool,
	}, nil
}

func MustConnectDB(config *config.Config, ctx context.Context) *DbConn {
	db, err := ConnectSqlcWithPool(config, ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}
