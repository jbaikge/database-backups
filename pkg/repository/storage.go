package repository

import (
	"database/sql"
	"time"

	"github.com/jbaikge/database-backups/pkg/api"
)

type Storage interface {
	CreateDatabase(api.NewDatabaseRequest) error
	CreateServer(api.NewServerRequest) error
	DeleteDatabase(int) error
	DeleteServer(int) error
	GetDatabase(int) (*api.Database, error)
	GetServer(int) (*api.Server, error)
	ListDatabases() ([]api.Database, error)
	ListServers() ([]api.Server, error)
	RunMigrations() error
	ServerTree() ([]api.Tree, error)
	UpdateDatabase(int, api.UpdateDatabaseRequest) error
	UpdateServer(int, api.NewServerRequest) error
}

type storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) Storage {
	return &storage{
		db: db,
	}
}

func (s *storage) CreateDatabase(db api.NewDatabaseRequest) error {
	query := `
		INSERT INTO databases (
			server_id,
			name,
			added
		) VALUES (
			$1,
			$2,
			$3
		)
	`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(db.ServerId, db.Name, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) CreateServer(server api.NewServerRequest) error {
	query := `
		INSERT INTO servers (
			name,
			host,
			port,
			username,
			password,
			proxy_host,
			proxy_username,
			proxy_identity
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		server.Name,
		server.Host,
		server.Port,
		server.Username,
		server.Password,
		server.ProxyHost,
		server.ProxyUsername,
		server.ProxyIdentity,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) DeleteDatabase(id int) error {
	query := `
		UPDATE databases SET backup = 0, removed = $1 WHERE database_id = $2
	`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) DeleteServer(id int) error {
	queryDatabase := `DELETE FROM databases WHERE server_id = $1`
	queryServer := `DELETE FROM servers WHERE server_id = $1`

	stmtDatabase, err := s.db.Prepare(queryDatabase)
	if err != nil {
		return err
	}
	defer stmtDatabase.Close()

	if _, err := stmtDatabase.Exec(id); err != nil {
		return err
	}

	stmtServer, err := s.db.Prepare(queryServer)
	if err != nil {
		return err
	}
	defer stmtServer.Close()

	if _, err := stmtServer.Exec(id); err != nil {
		return err
	}

	return nil
}

func (s *storage) GetDatabase(id int) (*api.Database, error) {
	query := `
		SELECT
			database_id,
			server_id,
			name,
			backup,
			only_tables,
			exclude_tables,
			added,
			removed
		FROM databases
		WHERE database_id = $1
	`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	db := new(api.Database)
	err = stmt.QueryRow(id).Scan(
		&db.Id,
		&db.ServerId,
		&db.Name,
		&db.Backup,
		&db.OnlyTables,
		&db.ExcludeTables,
		&db.Added,
		&db.Removed,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *storage) GetServer(id int) (*api.Server, error) {
	query := `
		SELECT
			server_id,
			name,
			host,
			port,
			username,
			password,
			proxy_host,
			proxy_username,
			proxy_identity
		FROM servers
		WHERE server_id = $1
		ORDER BY name ASC
	`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	server := new(api.Server)
	err = stmt.QueryRow(id).Scan(
		&server.Id,
		&server.Name,
		&server.Host,
		&server.Port,
		&server.Username,
		&server.Password,
		&server.ProxyHost,
		&server.ProxyUsername,
		&server.ProxyIdentity,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (s *storage) ListDatabases() ([]api.Database, error) {
	query := `
		SELECT
			database_id,
			server_id,
			name,
			backup,
			only_tables,
			exclude_tables,
			added,
			removed
		FROM databases
		ORDER BY name ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbs := make([]api.Database, 0, 1000)
	for rows.Next() {
		var db api.Database
		err := rows.Scan(
			&db.Id,
			&db.ServerId,
			&db.Name,
			&db.Backup,
			&db.OnlyTables,
			&db.ExcludeTables,
			&db.Added,
			&db.Removed,
		)
		if err != nil {
			return nil, err
		}
		dbs = append(dbs, db)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dbs, nil
}

func (s *storage) ListServers() ([]api.Server, error) {
	query := `
		SELECT
			server_id,
			name,
			host,
			port,
			username,
			password,
			proxy_host,
			proxy_username,
			proxy_identity
		FROM servers
		ORDER BY name ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := make([]api.Server, 0, 100)
	for rows.Next() {
		var v api.Server
		rows.Scan(
			&v.Id,
			&v.Name,
			&v.Host,
			&v.Port,
			&v.Username,
			&v.Password,
			&v.ProxyHost,
			&v.ProxyUsername,
			&v.ProxyIdentity,
		)
		servers = append(servers, v)
	}
	return servers, nil
}

func (s *storage) RunMigrations() error {
	for _, migration := range migrations {
		run, err := migration.check(s.db)
		if err != nil {
			return err
		}
		if !run {
			continue
		}
		if _, err := s.db.Exec(migration.sql); err != nil {
			return err
		}
	}
	return nil
}

func (s *storage) ServerTree() ([]api.Tree, error) {
	servers, err := s.ListServers()
	if err != nil {
		return nil, err
	}

	databases, err := s.ListDatabases()
	if err != nil {
		return nil, err
	}

	trees := make([]api.Tree, len(servers))
	for i, server := range servers {
		trees[i].Server = server
		trees[i].Databases = make([]api.Database, 0, len(databases))
		for _, db := range databases {
			if db.ServerId == server.Id {
				trees[i].Databases = append(trees[i].Databases, db)
			}
		}
	}

	return trees, nil
}

func (s *storage) UpdateDatabase(id int, db api.UpdateDatabaseRequest) error {
	query := `
		UPDATE databases SET
			server_id      = $1,
			name           = $2,
			backup         = $3,
			only_tables    = $4,
			exclude_tables = $5
		WHERE database_id = $6
	`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		db.ServerId,
		db.Name,
		db.Backup,
		db.OnlyTables,
		db.ExcludeTables,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) UpdateServer(id int, server api.NewServerRequest) error {
	query := `
		UPDATE servers SET
			name           = $1,
			host           = $2,
			port           = $3,
			username       = $4,
			password       = $5,
			proxy_host     = $6,
			proxy_username = $7,
			proxy_identity = $8
		WHERE server_id = $9
	`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		server.Name,
		server.Host,
		server.Port,
		server.Username,
		server.Password,
		server.ProxyHost,
		server.ProxyUsername,
		server.ProxyIdentity,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}
