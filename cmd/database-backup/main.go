package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/jbaikge/database-backups/pkg/api"
	"github.com/jbaikge/database-backups/pkg/repository"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %s\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	databasePath := "/tmp/database-backups.sqlite3"
	dumpDir := "/tmp/dumps"
	onlyUpdateList := false

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&databasePath, "db", databasePath, "Path to configuration and logging database")
	flags.StringVar(&dumpDir, "dir", dumpDir, "Directory to store dumps")
	flags.BoolVar(&onlyUpdateList, "update-list", onlyUpdateList, "Only update database lists for servers")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	db, err := setupDatabase(databasePath)
	if err != nil {
		return err
	}

	storage := repository.NewStorage(db)
	if err := storage.RunMigrations(); err != nil {
		return err
	}

	serverService := api.NewServerService(storage)
	databaseService := api.NewDatabaseService(storage)

	servers, err := serverService.List()
	if err != nil {
		return err
	}

	for _, server := range servers {
		if err := updateDatabaseList(server, databaseService); err != nil {
			return err
		}
	}

	if onlyUpdateList {
		return nil
	}

	if err := os.MkdirAll(dumpDir, 0755); err != nil {
		return err
	}

	return nil
}

func setupDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func buildServerCmd(server api.Server, cmd string, args ...string) ([]string, error) {
	parts := make([]string, 0, 16)
	if server.ProxyHost != "" {
		userHost := fmt.Sprintf("%s@%s", server.ProxyUsername, server.ProxyHost)
		parts = append(parts, "ssh", "-i", server.ProxyIdentity, userHost)
	}
	parts = append(parts, cmd, "-h", server.Host, "-u", server.Username)
	if server.Password != "" {
		password, err := server.DecryptPassword()
		if err != nil {
			return nil, err
		}
		parts = append(parts, fmt.Sprintf("-p%s", password))
	}
	parts = append(parts, args...)
	return parts, nil
}

func databaseList(server api.Server) ([]string, error) {
	parts, err := buildServerCmd(
		server,
		"mysql",
		"--skip-column-names",
		"--batch",
		"--execute",
		"SHOW DATABASES WHERE `Database` NOT IN('mysql', 'information_schema', 'performance_schema')",
	)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Fields(string(output)), nil
}

// func dumpDatabase(server api.Server, database api.Database) error {
// 	return nil
// }

func updateDatabaseList(server api.Server, databaseService api.DatabaseService) error {
	log.Printf("Checking %s for new databases", server.Name)
	existing, err := databaseService.ListServer(server.Id)
	if err != nil {
		return err
	}

	existingNames := make([]string, len(existing))
	for i, db := range existing {
		existingNames[i] = db.Name
	}

	// Database list is already sorted by name, but verify for the sake of
	// the following steps
	if !sort.StringsAreSorted(existingNames) {
		return errors.New("database names are not sorted")
	}

	names, err := databaseList(server)
	if err != nil {
		return err
	}

	// Add in new databases
	for _, name := range names {
		if x := sort.SearchStrings(existingNames, name); x < len(existingNames) && existingNames[x] == name {
			continue
		}

		log.Printf("New database on %s: %s", server.Name, name)
		newDb := api.NewDatabaseRequest{
			Name:     name,
			ServerId: server.Id,
		}
		if err := databaseService.New(newDb); err != nil {
			return err
		}
	}

	// Mark removed databases
	for _, db := range existing {
		if x := sort.SearchStrings(names, db.Name); x < len(names) && names[x] == db.Name {
			continue
		}

		if !db.Backup && db.Removed != nil {
			continue
		}

		log.Printf("Removing database from %s: %s", server.Name, db.Name)
		if err := databaseService.Delete(db.Id); err != nil {
			return err
		}
	}
	return nil
}