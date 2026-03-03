package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseConfig struct {
	Port         string `env:"DB_PORT" envDefault:"5432"`
	Host         string `env:"DB_HOST" envDefault:"localhost"`
	User         string `env:"DB_USER" envDefault:"postgres"`
	Password     string `env:"DB_PASSWORD" envDefault:"postgres"`
	DatabaseName string `env:"DB_NAME" envDefault:"librarydb"`
}

func (conf databaseConfig) databaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", conf.User, conf.Password, conf.Host, conf.Port, conf.DatabaseName)
}

type LibraryDatabase struct {
	pool     *pgxpool.Pool
	dbConfig databaseConfig
}

func NewLibraryDatabase() *LibraryDatabase {
	dbConfig := databaseConfig{}
	if err := env.Parse(&dbConfig); err != nil {
		log.Printf("Failed to parse database config from environment: %v", err)
	}
	return &LibraryDatabase{
		dbConfig: dbConfig,
	}
}

func (database *LibraryDatabase) Connect() error {
	poolConfig, err := pgxpool.ParseConfig(database.dbConfig.databaseURL())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse DATABASE_URL: %s\n", err)
		return err
	}

	database.pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to create connection pool:", err)
	}
	return nil
}

func (database *LibraryDatabase) Close() {
	if database.pool != nil {
		database.pool.Close()
	}
}

func (database *LibraryDatabase) Exec(ctx context.Context, s string, i ...interface{}) (pgconn.CommandTag, error) {
	return database.pool.Exec(ctx, s, i...)
}

func (database *LibraryDatabase) Query(ctx context.Context, s string, i ...interface{}) (pgx.Rows, error) {
	return database.pool.Query(ctx, s, i...)
}

func (database *LibraryDatabase) QueryRow(ctx context.Context, s string, i ...interface{}) pgx.Row {
	return database.pool.QueryRow(ctx, s, i...)
}

func (database *LibraryDatabase) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return database.pool.BeginTx(ctx, opts)
}

func (database *LibraryDatabase) ExecTx(ctx context.Context, tx pgx.Tx, s string, i ...interface{}) (pgconn.CommandTag, error) {
	return tx.Exec(ctx, s, i...)
}
