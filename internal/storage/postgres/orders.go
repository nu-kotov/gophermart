package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/nu-kotov/gophermart/internal/storage/dberrors"
)

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

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			if strings.Contains(pgErr.Message, "orders_pkey") {
				return dberrors.ErrUserOrderDuplicate
			}

			return dberrors.ErrOrderDuplicate
		}

		return err
	}

	return tx.Commit()
}

func (pg *DBStorage) SelectOrdersByUserID(ctx context.Context, userID string) ([]models.GetUserOrdersResponse, error) {
	var data []models.GetUserOrdersResponse

	query := `SELECT number, status, accrual, uploaded_at from orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := pg.db.Query(query, userID)

	if err != nil {
		return nil, dberrors.ErrNotFound
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

		data = append(data, models.GetUserOrdersResponse{
			Number:     strconv.FormatInt(number, 10),
			Status:     status,
			Accrual:    accrual,
			UploadedAt: uploadedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func (pg *DBStorage) UpdateOrder(ctx context.Context, pointsData *models.OrderData) error {

	updateOrder := `UPDATE orders SET status=$1, accrual=$2 WHERE number=$3`
	currentBalance := `SELECT balance FROM users_balances WHERE user_id=$1`
	updateUsersBalances := `
	    INSERT INTO users_balances (balance, user_id) VALUES ($1, $2) ON CONFLICT (user_id)
	    DO UPDATE 
		    SET balance=$1;
	`

	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		updateOrder,
		pointsData.Status,
		pointsData.Accrual,
		pointsData.Number,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	row := pg.db.QueryRowContext(
		ctx,
		currentBalance,
		pointsData.UserID,
	)

	var curBalance float64
	err = row.Scan(&curBalance)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			curBalance = 0.0
		} else {
			return err
		}
	}

	_, err = tx.ExecContext(
		ctx,
		updateUsersBalances,
		curBalance+pointsData.Accrual,
		pointsData.UserID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (pg *DBStorage) SelectUnprocessedOrders(ctx context.Context) ([]models.OrderData, error) {
	var unprocessedOrders []models.OrderData

	query := `SELECT number, user_id, status, accrual FROM orders WHERE status IN ('NEW', 'REGISTERED', 'PROCESSING') ORDER BY uploaded_at DESC`

	rows, err := pg.db.Query(query)

	if err != nil {
		return nil, dberrors.ErrNotFound
	}

	for rows.Next() {
		var number int64
		var userID string
		var accrual float64
		var status string

		err := rows.Scan(&number, &userID, &status, &accrual)

		if err != nil {
			return nil, err
		}

		unprocessedOrders = append(unprocessedOrders, models.OrderData{
			Number:  number,
			Status:  status,
			Accrual: accrual,
			UserID:  userID,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return unprocessedOrders, nil
}
