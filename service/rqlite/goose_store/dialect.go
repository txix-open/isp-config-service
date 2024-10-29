package goose_store

import (
	"fmt"
)

type Rqlite struct{}

func (s *Rqlite) CreateTable(tableName string) string {
	q := `CREATE TABLE if not exists %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version_id INTEGER NOT NULL,
		is_applied INTEGER NOT NULL,
		tstamp DATETIME DEFAULT (datetime('now'))
	)`
	return fmt.Sprintf(q, tableName)
}

func (s *Rqlite) InsertVersion(tableName string) string {
	q := `INSERT INTO %s (version_id, is_applied) VALUES (?, ?)`
	return fmt.Sprintf(q, tableName)
}

func (s *Rqlite) DeleteVersion(tableName string) string {
	q := `DELETE FROM %s WHERE version_id=?`
	return fmt.Sprintf(q, tableName)
}

func (s *Rqlite) GetMigrationByVersion(tableName string) string {
	q := `SELECT tstamp, is_applied FROM %s WHERE version_id=? ORDER BY tstamp DESC LIMIT 1`
	return fmt.Sprintf(q, tableName)
}

func (s *Rqlite) ListMigrations(tableName string) string {
	q := `SELECT version_id, is_applied from %s ORDER BY id DESC`
	return fmt.Sprintf(q, tableName)
}
