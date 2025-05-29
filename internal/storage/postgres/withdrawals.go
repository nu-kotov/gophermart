package postgres

import (
	"context"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/gophermart/internal/models"
)

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
