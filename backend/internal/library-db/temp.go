package library_db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alexsieland/bg-library/api"
	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgxutil"
)

var conn *pgx.Conn

type dbConfig struct {
	Port         string `env:"PGSQL_PORT" envDefault:"5432"`
	Host         string `env:"PGSQL_HOST" envDefault:"localhost"`
	User         string `env:"PGSQL_USER" envDefault:"postgres"`
	Password     string `env:"PGSQL_PASSWORD" envDefault:"postgres"`
	DatabaseName string `env:"PGSQL_DATABASE" envDefault:"librarydb"`
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

func AddGame(title string) (api.Game, error) {
	//Clean inputs
	trimmedTitle := strings.TrimSpace(title)

	//Validate inputs
	if len(trimmedTitle) > 100 {
		return api.Game{}, fmt.Errorf("title cannot be longer than 100 characters")
	}

	if len(trimmedTitle) == 0 {
		return api.Game{}, fmt.Errorf("title cannot be blank")
	}

	//Make call
	pgxutil.InsertRowReturning(context.Background(), pgxutil.Queryer())
	result, err := conn.QueryRow(context.Background(), "INSERT INTO games (title) VALUES ($1)", trimmedTitle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to insert game in database: %v\n", err)
		return api.Game{}, err
	}

	result.

}
