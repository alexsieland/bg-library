package db

import (
	"context"
	"fmt"
	"log"
	"os"

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

var config = databaseConfig{}

func (conf databaseConfig) databaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", conf.User, conf.Password, conf.Host, conf.Port, conf.DatabaseName)
}

type LibraryDatabase struct {
	pool     *pgxpool.Pool
	dbConfig databaseConfig
}

func NewLibraryDatabase() LibraryDatabase {
	return LibraryDatabase{
		dbConfig: databaseConfig{},
	}
}

func (database LibraryDatabase) Connect() error {
	poolConfig, err := pgxpool.ParseConfig(config.databaseURL())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse DATABASE_URL: %s\n", err)
		return err
	}

	database.pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to create connection pool:", err)
	}

	defer database.pool.Close()
	return nil
}

func (database LibraryDatabase) Close() {
	if database.pool != nil {
		database.pool.Close()
	}
}

func (database LibraryDatabase) Exec(ctx context.Context, s string, i ...interface{}) (pgconn.CommandTag, error) {
	return database.pool.Exec(ctx, s, i...)
}

func (database LibraryDatabase) Query(ctx context.Context, s string, i ...interface{}) (pgx.Rows, error) {
	return database.pool.Query(ctx, s, i...)
}

func (database LibraryDatabase) QueryRow(ctx context.Context, s string, i ...interface{}) pgx.Row {
	return database.pool.QueryRow(ctx, s, i...)
}
