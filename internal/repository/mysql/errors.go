package mysql

import "github.com/go-sql-driver/mysql"

func isMySQLDuplicateEntry(err error) bool {
	if e, ok := err.(*mysql.MySQLError); ok {
		return e.Number == 1062
	}
	return false
}
