package api

import (
	"fmt"
	"path/filepath"
	"time"
)

func (s Server) Filename(d Database) string {
	return fmt.Sprintf("%s_%s_%s.sql", s.Name, d.Name, time.Now().Format("2006-01-02"))

}

func (s Server) S3Key(d Database) string {
	return filepath.Join(s.Name, d.Name, s.Filename(d))
}
