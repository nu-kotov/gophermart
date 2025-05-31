package postgres

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/gophermart/internal/models"
)

type UsersStorage struct {
	Stor *DBStorage
}

func (usrs *UsersStorage) InsertUserData(ctx context.Context, data *models.UserData) error {

	sql := `INSERT INTO users (login, password) VALUES ($1, $2);`

	tx, err := usrs.Stor.db.Begin()
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

func (usrs *UsersStorage) SelectUserData(ctx context.Context, data *models.UserData) (*models.UserData, error) {

	var userData models.UserData

	sql := `SELECT user_id, password from users WHERE login = $1`

	row := usrs.Stor.db.QueryRowContext(
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
