package repository

import "database/sql"

type checkFunc func(*sql.DB) (bool, error)

type migration struct {
	sql   string
	check checkFunc
}

var migrations = []migration{
	{
		sql: `
			CREATE TABLE servers (
				server_id      INTEGER PRIMARY KEY,
				name           TEXT NOT NULL,
				host           TEXT NOT NULL,
				port           TEXT NOT NULL,
				username       TEXT NOT NULL,
				password       TEXT NOT NULL,
				proxy_host     TEXT,
				proxy_username TEXT,
				proxy_identity TEXT
			)
		`,
		check: checkTableExists("servers"),
	},
	{
		sql: `
			CREATE TABLE databases (
				database_id    INTEGER PRIMARY KEY,
				server_id      INTEGER NOT NULL,
				name           TEXT NOT NULL,
				only_tables    TEXT NOT NULL DEFAULT '', -- Space-separated list
				exclude_tables TEXT NOT NULL DEFAULT '', -- Space-separated list
				backup         INTEGER NOT NULL DEFAULT 1,
				added          DATETIME NOT NULL,
				removed        DATETIME,
				UNIQUE (server_id, name),
				FOREIGN KEY (server_id) REFERENCES servers (server_id)
			)
		`,
		check: checkTableExists("databases"),
	},
	{
		sql: `
			CREATE TABLE logs (
				log_id        INTEGER PRIMARY KEY,
				database_id   INTEGER NOT NULL,
				backup_start  DATETIME NULL,
				backup_end    DATETIME NULL,
				size_previous INTEGER NOT NULL DEFAULT 0,
				size_current  INTEGER NOT NULL DEFAULT 0,
				added         DATETIME NOT NULL,
				FOREIGN KEY (database_id) REFERENCES servers (database_id)
			)
		`,
		check: checkTableExists("logs"),
	},
}

func checkTableExists(name string) checkFunc {
	return func(db *sql.DB) (bool, error) {
		sql := `
			SELECT name
			FROM sqlite_master
			WHERE
				type = 'table'
				AND name = ?
		`
		stmt, err := db.Prepare(sql)
		if err != nil {
			return false, err
		}
		rows, err := stmt.Query(name)
		if err != nil {
			return false, err
		}
		defer rows.Close()
		return !rows.Next(), nil
	}
}
