package mysql

import "github.com/go-sql-driver/mysql"

func isMySQLDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	mysqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	return mysqlErr.Number == 1062
}
