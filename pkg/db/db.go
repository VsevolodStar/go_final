package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Переменная для хранения глобального подключения к БД.
var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL,
	title VARCHAR NOT NULL,
	comment TEXT,
	repeat VARCHAR(128)
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init открывает подключение к БД и создаёт ее по схеме, при необходимости
func Init(dbFile string) error {
	d, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if _, err := d.Exec(schema); err != nil {
		d.Close()
		return err
	}

	DB = d
	return nil
}
