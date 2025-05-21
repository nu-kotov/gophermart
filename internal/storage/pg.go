package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/pressly/goose/v3"
)

var ErrUserOrderDuplicate = errors.New("current user data conflict")
var ErrOrderDuplicate = errors.New("data conflict")
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

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			if strings.Contains(pgErr.Message, "orders_pkey") {
				return ErrUserOrderDuplicate
			}

			return ErrOrderDuplicate
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

func (pg *DBStorage) UpdateUserBalance(ctx context.Context, newBalance *models.UserBalance, withdraw *models.Withdraw) error {

	updateUsersBalances := `UPDATE users_balances SET balance=$1, withdrawn=$2 WHERE user_id = $3`
	insertWithdrawal := `INSERT INTO withdrawals (number, user_id, sum, withdrawn_at) VALUES ($1, $2, $3, $4);`

	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		updateUsersBalances,
		newBalance.Balance,
		newBalance.Withdrawn,
		withdraw.UserID,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		insertWithdrawal,
		withdraw.Number,
		withdraw.UserID,
		withdraw.Sum,
		withdraw.WithdrawnAt,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (pg *DBStorage) SelectUserWithdrawals(ctx context.Context, userID string) ([]models.WithdrawnInfo, error) {
	var data []models.WithdrawnInfo

	query := `SELECT number, sum, withdrawn_at FROM withdrawals WHERE user_id = $1 ORDER BY withdrawn_at DESC`

	rows, err := pg.db.Query(query, userID)

	if err != nil {
		return nil, ErrNotFound
	}

	for rows.Next() {
		var number int64
		var sum float64
		var withdrawnAt time.Time

		err := rows.Scan(&number, &sum, &withdrawnAt)

		if err != nil {
			return nil, err
		}

		data = append(data, models.WithdrawnInfo{
			Number:      strconv.FormatInt(number, 10),
			Sum:         sum,
			WithdrawnAt: withdrawnAt.Format(time.RFC1123),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

// func (pg *DBStorage) SelectUnprocessedOrders(ctx context.Context) ([]string, error) {
// 	var numbers []string

// 	query := `SELECT number FROM orders WHERE status IN ('REGISTERED', 'PROCESSING') ORDER BY uploaded_at DESC`

// 	rows, err := pg.db.Query(query)

// 	if err != nil {
// 		return nil, ErrNotFound
// 	}

// 	for rows.Next() {
// 		var number int64

// 		err := rows.Scan(&number)

// 		if err != nil {
// 			return nil, err
// 		}

// 		fmt.Println(err)

// 		numbers = append(numbers, strconv.FormatInt(number, 10))
// 	}
// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return numbers, nil
// }

func (pg *DBStorage) UpdateOrder(ctx context.Context, pointsData *models.Orders) error {

	updateOrder := `UPDATE orders SET status=$1, accrual=$2 WHERE number=$3`
	updateUsersBalances := `UPDATE users_balances SET balance=balance+$1 WHERE user_id=$2`

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

	_, err = tx.ExecContext(
		ctx,
		updateUsersBalances,
		pointsData.Accrual,
		pointsData.UserID,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (pg *DBStorage) SelectUnprocessedOrders(ctx context.Context) ([]models.Orders, error) {
	var unprocessedOrders []models.Orders

	query := `SELECT number, user_id, status, accrual FROM orders WHERE status IN ('REGISTERED', 'PROCESSING') ORDER BY uploaded_at DESC`

	rows, err := pg.db.Query(query)

	if err != nil {
		return nil, ErrNotFound
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

		unprocessedOrders = append(unprocessedOrders, models.Orders{
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
