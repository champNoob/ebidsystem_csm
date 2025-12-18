package database

import (
	"database/sql"

	"ebidsystem_csm/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

var MySQL *sql.DB

func InitMySQL(cfg config.MySQLConfig) error {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	MySQL = db
	return nil
}
