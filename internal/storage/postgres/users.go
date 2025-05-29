package postgres

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/gophermart/internal/models"
)

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

func (pg *DBStorage) SelectUserData(ctx context.Context, data *models.UserData) (*models.UserData, error) {

	var userData models.UserData

	sql := `SELECT user_id, password from users WHERE login = $1`

	row := pg.db.QueryRowContext(
		ctx,
		sql,
		data.Login,
	)

	err := row.Scan(&userData.UserID, &userData.Password)
	if err != nil {
		return nil, err
	}

	return &userData, nil
}
