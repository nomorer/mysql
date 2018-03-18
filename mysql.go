package mysql

import (
	"database/sql"
)

type Mysql struct {
	*sql.DB
}

func NewMysql(db *sql.DB) *Mysql {
	return &Mysql{db}
}

func (mysql *Mysql) QueryRow(v interface{}, query string, args ...interface{}) error {
	return mysql.queryRows(func(rows *sql.Rows) error {
		return UnmarshalRow(v, rows)
	}, query, args...)
}

func (mysql *Mysql) QueryRows(v interface{}, query string, args ...interface{}) error {
	return mysql.queryRows(func(rows *sql.Rows) error {
		return UnmarshalRows(v, rows)
	}, query, args...)
}

func (mysql *Mysql) queryRows(scanner func(*sql.Rows) error, query string, args ...interface{}) error {
	rows, err := mysql.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanner(rows)
}
