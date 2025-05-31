package postgres

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/nu-kotov/gophermart/internal/storage/dberrors"
)

type BalanceStorage struct {
	Stor *DBStorage
}

func (bs *BalanceStorage) SelectUserBalance(ctx context.Context, userID string) (*models.UserBalance, error) {

	var userBalance models.UserBalance

	query := `SELECT balance, withdrawn FROM users_balances WHERE user_id = $1`

	row := bs.Stor.db.QueryRowContext(
		ctx,
		query,
		userID,
	)

	err := row.Scan(&userBalance.Current, &userBalance.Withdrawn)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, dberrors.ErrUserNoBalance
		}

		return nil, err
	}

	return &userBalance, nil
}

func (bs *BalanceStorage) UpdateUserBalance(ctx context.Context, newBalance *models.UserBalance, withdraw *models.Withdraw) error {

	updateUsersBalances := `UPDATE users_balances SET balance=$1, withdrawn=$2 WHERE user_id = $3`
	insertWithdrawal := `INSERT INTO withdrawals (number, user_id, sum, withdrawn_at) VALUES ($1, $2, $3, $4);`

	tx, err := bs.Stor.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		updateUsersBalances,
		newBalance.Current,
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
