package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

var conn *pgx.Conn

type dbConfig struct {
	Port         string `env:"DB_PORT" envDefault:"5432"`
	Host         string `env:"DB_HOST" envDefault:"localhost"`
	User         string `env:"DB_USER" envDefault:"postgres"`
	Password     string `env:"DB_PASSWORD" envDefault:"postgres"`
	DatabaseName string `env:"DB_NAME" envDefault:"librarydb"`
}

var config = dbConfig{}

func (conf dbConfig) databaaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", conf.User, conf.Password, conf.Host, conf.Port, conf.DatabaseName)
}

func Connect() error {
	c, err := pgx.Connect(context.Background(), config.databaaseURL())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return err
	}
	conn = c
	defer conn.Close(context.Background())
	return nil
}

func Close() error {
	if conn != nil {
		err := conn.Close(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to close connection to database: %v\n", err)
			return err
		}
	}
	return nil
}
