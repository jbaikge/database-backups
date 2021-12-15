package repository_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/zeebo/assert"
	"github.com/jbaikge/database-backups/pkg/api"
	"github.com/jbaikge/database-backups/pkg/repository"

	_ "github.com/mattn/go-sqlite3"
)

func TestDatabaseDateTime(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.Nil(t, err)

	storage := repository.NewStorage(db)
	assert.Nil(t, storage.RunMigrations())

	newDb := api.NewDatabaseRequest{
		Name: "test",
	}
	assert.Nil(t, storage.CreateDatabase(newDb))

	getDb, err := storage.GetDatabase(1)
	assert.Nil(t, err)
	assert.That(t, getDb.Name == "test")
	assert.False(t, getDb.Added.IsZero())
}

func TestDateTime(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.Nil(t, err)

	_, err = db.Exec(`CREATE TABLE a (id INTEGER PRIMARY KEY, d DATETIME)`)
	assert.Nil(t, err)

	stmt, err := db.Prepare(`INSERT INTO a (d) VALUES ($1)`)
	assert.Nil(t, err)
	defer stmt.Close()

	_, err = stmt.Exec(time.Now())
	assert.Nil(t, err)

	var row struct {
		Id   int
		Date time.Time
	}
	err = db.QueryRow(`SELECT id, d FROM a WHERE id = 1`).Scan(
		&row.Id,
		&row.Date,
	)
	assert.Nil(t, err)
	assert.False(t, row.Date.IsZero())
}
