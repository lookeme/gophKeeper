package utils

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	SettingKeyKek string = "setting_kek"
	SettingKeyDek string = "setting_dek"
)

// ErrorCode - compare errors
func ErrorCode(err error) string {
	var pgerr *pgconn.PgError
	ok := errors.As(err, &pgerr)
	if !ok {
		return ""
	}
	return pgerr.Code
}
