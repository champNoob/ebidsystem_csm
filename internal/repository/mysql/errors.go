package mysql

import (
	"ebidsystem_csm/internal/apperror"

	driver_mysql "github.com/go-sql-driver/mysql"
)

func isMySQLDuplicateEntry(err error) bool {
	if e, ok := err.(*driver_mysql.MySQLError); ok {
		return e.Number == 1062
	}
	return false
}

func isMySQLDeadlock(err error) bool {
	if e, ok := err.(*driver_mysql.MySQLError); ok {
		return e.Number == 1213
	}
	return false
}

func isMySQLLockWaitTimeout(err error) bool {
	if e, ok := err.(*driver_mysql.MySQLError); ok {
		return e.Number == 1205
	}
	return false
}

func wrapDBError(err error) error {
	if err == nil {
		return nil
	}

	// 数据库的底层错误不应暴露给前端。cause 会保留在错误链中，但对外只显示 ErrInternal.Message
	return apperror.Wrap(apperror.ErrInternal, err)
}
