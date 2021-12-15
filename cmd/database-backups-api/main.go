package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/jbaikge/database-backups/pkg/api"
	"github.com/jbaikge/database-backups/pkg/app"
	"github.com/jbaikge/database-backups/pkg/repository"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Start-up error: %s\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	databasePath := "/tmp/database-backups.sqlite3"
	listenAddress := "0.0.0.0:3000"

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&databasePath, "db", databasePath, "Path to configuration and logging database")
	flags.StringVar(&listenAddress, "addr", listenAddress, "API listening address")
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

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(cors.Default())

	serverService := api.NewServerService(storage)
	databaseService := api.NewDatabaseService(storage)

	server := app.NewServer(router, serverService, databaseService)

	return server.Run(listenAddress)
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
