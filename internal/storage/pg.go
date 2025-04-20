package storage

import (
	"context"
	"database/sql"
	"embed"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/pressly/goose/v3"
)

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

func (pg *DBStorage) InsertUserData(ctx context.Context, data *models.UserData) error {

	sql := `INSERT INTO users (login, password) VALUES ($1, $2);`

	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		sql,
		data.Login,
		data.Password,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
