package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	onlyUpdate := false
	bucket := ""

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&databasePath, "db", databasePath, "Path to configuration and logging database")
	flags.StringVar(&dumpDir, "dir", dumpDir, "Directory to store dumps")
	flags.StringVar(&bucket, "bucket", bucket, "AWS Bucket to store dumps")
	flags.BoolVar(&onlyUpdate, "update", onlyUpdate, "Only update database lists for servers")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	// Ensure bucket is set as I keep forgetting to set this flag
	if bucket == "" {
		return errors.New("bucket name is empty")
	}

	// Check for required environment variables
	envvars := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_REGION",
	}
	for _, envvar := range envvars {
		if os.Getenv(envvar) == "" {
			return errors.New("environment variable, " + envvar + " is required")
		}
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
		log.Printf("Updating database list for %s", server.Name)
		if err := serverService.UpdateDatabases(server.Id); err != nil {
			return err
		}
	}

	// Bail if we are only updating the database lists
	if onlyUpdate {
		return nil
	}

	if err := os.MkdirAll(dumpDir, 0755); err != nil {
		return err
	}

	for _, server := range servers {
		log.Printf("Dumping databases in %s", server.Name)
		databases, err := databaseService.List(server.Id)
		if err != nil {
			return err
		}
		for _, database := range databases {
			if !database.Backup {
				continue
			}

			filename := filepath.Join(dumpDir, server.Filename(database))
			log.Printf("Dumping %s to %s", database.Name, filename)
			if err := dumpDatabase(server, database, filename); err != nil {
				return err
			}

			key := server.S3Key(database)
			log.Printf("Sending to S3 s3://%s/%s", bucket, key)
			if err := sendToS3(bucket, filename, key); err != nil {
				return err
			}

			log.Printf("Removing temporary file")
			if err := os.Remove(filename); err != nil {
				return err
			}
		}
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

func dumpDatabase(server api.Server, database api.Database, path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// In case this runs twice in the same day, empty the file before writing
	if err := file.Truncate(0); err != nil {
		return err
	}

	args, err := server.DatabaseDumpCmd(database)
	if err != nil {
		return err
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = file
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// Required environment variables:
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
// AWS_REGION
func sendToS3(bucket string, filename string, key string) error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	input := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	}

	uploader := s3manager.NewUploader(sess)
	if _, err := uploader.Upload(input); err != nil {
		return err
	}
	return nil
}
