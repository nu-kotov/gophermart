package postgres

import (
	"database/sql"
	"embed"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var ErrUserOrderDuplicate = errors.New("current user data conflict")
var ErrOrderDuplicate = errors.New("data conflict")
var ErrNotFound = errors.New("data not found")
var ErrUserNoBalance = errors.New("user have not balance")

type DBStorage struct {
	db *sql.DB
}

var (
	dbInstance *DBStorage
	//go:embed migrations/*.sql
	embedMigrations embed.FS
)

func NewConnect(connString string) (*DBStorage, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}

	dbInstance = &DBStorage{db}

	return dbInstance, nil
}

func (pg *DBStorage) Close() error {
	return pg.db.Close()
}
