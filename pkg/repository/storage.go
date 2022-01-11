package repository

import (
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/jbaikge/database-backups/pkg/api"
)

type Storage interface {
	CreateDatabase(api.NewDatabaseRequest) error
	CreateServer(api.NewServerRequest) (int, error)
	DeleteDatabase(int) error
	DeleteServer(int) error
	GetDatabase(int) (*api.Database, error)
	GetServer(int) (*api.Server, error)
	ListDatabases(int) ([]api.Database, error)
	ListServers() ([]api.Server, error)
	RunMigrations() error
	ServerTree() ([]api.Tree, error)
	UpdateDatabase(int, api.UpdateDatabaseRequest) error
	UpdateServer(int, api.NewServerRequest) error
	UpdateServerDatabases(int, []string) error
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

func (s *storage) CreateServer(server api.NewServerRequest) (id int, err error) {
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
		return
	}

	result, err := stmt.Exec(
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
		return
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return
	}

	id = int(id64)
	return
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

func (s *storage) ListDatabases(serverId int) ([]api.Database, error) {
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
		WHERE server_id = $1
		ORDER BY name ASC
	`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(serverId)
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

	trees := make([]api.Tree, len(servers))
	for i, server := range servers {
		trees[i].Server = server
		trees[i].Databases, err = s.ListDatabases(server.Id)
		if err != nil {
			return nil, err
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

func (s *storage) UpdateServerDatabases(id int, names []string) error {
	existing, err := s.ListDatabases(id)
	if err != nil {
		return err
	}

	existingNames := make([]string, len(existing))
	for i, db := range existing {
		existingNames[i] = db.Name
	}

	// Database list is already sorted by name, but verify for the sake of the
	// following steps
	if !sort.StringsAreSorted(existingNames) {
		return errors.New("database names are not sorted")
	}

	// Add new databases
	for _, name := range names {
		// Skip names that already exist
		if x := sort.SearchStrings(existingNames, name); x < len(existingNames) && existingNames[x] == name {
			continue
		}

		newDb := api.NewDatabaseRequest{
			Name:     name,
			ServerId: id,
		}
		if err := s.CreateDatabase(newDb); err != nil {
			return err
		}
	}

	// Mark removed databases
	for _, db := range existing {
		// Skip existing databases that are incoming
		if x := sort.SearchStrings(names, db.Name); x < len(names) && names[x] == db.Name {
			continue
		}

		// Skip databases already marked for removal
		if !db.Backup && db.Removed != nil {
			continue
		}

		if err := s.DeleteDatabase(db.Id); err != nil {
			return err
		}
	}
	return nil
}
