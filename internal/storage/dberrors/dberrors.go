package dberrors

import (
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrUserOrderDuplicate = errors.New("current user data conflict")
var ErrOrderDuplicate = errors.New("data conflict")
var ErrNotFound = errors.New("data not found")
var ErrUserNoBalance = errors.New("user have not balance")
