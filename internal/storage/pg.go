package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/pressly/goose/v3"
)

var ErrNotFound = errors.New("data not found")

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

func (pg *DBStorage) InsertOrderData(ctx context.Context, data *models.OrderData) error {

	sql := `INSERT INTO orders (number, user_id, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5);`

	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		sql,
		data.Number,
		data.UserID,
		data.Status,
		data.Accrual,
		data.UploadedAt,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (pg *DBStorage) SelectOrdersByUserID(ctx context.Context, userID string) ([]models.GetUserOrdersResponse, error) {
	var data []models.GetUserOrdersResponse

	query := `SELECT number, status, accrual, uploaded_at from orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := pg.db.Query(query, userID)

	if err != nil {
		return nil, ErrNotFound
	}

	for rows.Next() {
		var number int64
		var accrual float64
		var status string
		var uploadedAt time.Time

		err := rows.Scan(&number, &status, &accrual, &uploadedAt)

		if err != nil {
			return nil, err
		}

		fmt.Println(err)

		data = append(data, models.GetUserOrdersResponse{
			Number:     number,
			Status:     status,
			Accrual:    accrual,
			UploadedAt: uploadedAt.Format(time.RFC1123),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func (pg *DBStorage) SelectUserBalance(ctx context.Context, userID string) (*models.UserBalance, error) {

	var userBalance models.UserBalance

	sql := `SELECT balance, withdrawn from users_balances WHERE user_id = $1`

	row := pg.db.QueryRowContext(
		ctx,
		sql,
		userID,
	)

	err := row.Scan(&userBalance.Balance, &userBalance.Withdrawn)
	if err != nil {
		return nil, err
	}

	return &userBalance, nil
}

func (pg *DBStorage) UpdateUserBalance(ctx context.Context, userID string, newBalance *models.UserBalance) error {

	sql := `UPDATE users_balances SET balance=$1, withdrawn=$2 WHERE user_id = $3`

	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		sql,
		newBalance.Balance,
		newBalance.Withdrawn,
		userID,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
