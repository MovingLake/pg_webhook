package providers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
)

type Task struct {
	Ctx     context.Context
	Schema  string
	Table   string
	Verb    string
	Values  map[string]interface{}
	Retries int
}

type DbHandler interface {
	OpenDB()
	InsertTask(t Task)
	Close()
}

type SqliteHandler struct {
	db *sql.DB
}

var sqliteLocation = os.Getenv("SQLITE_LOCATION")

func (s SqliteHandler) Close() {
	s.db.Close()
}

func (s *SqliteHandler) OpenDB() {
	if sqliteLocation == "" {
		log.Println("SQLITE_LOCATION not set, using default location")
		sqliteLocation = "db.sqlite"
	}
	newdb, err := sql.Open("sqlite3", sqliteLocation)

	if err != nil {
		log.Fatal(err)
	}
	s.db = newdb
}

func (s SqliteHandler) InsertTask(t Task) {
	stmt, err := s.db.Prepare("CREATE TABLE IF NOT EXISTS tasks (schema TEXT, table_name TEXT, verb TEXT, values_table TEXT, retries INT)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}
	stmt.Close()

	stmt, err = s.db.Prepare("INSERT INTO tasks (schema, table_name, verb, values_table, retries) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Neeed to convert map to json string
	outBytes, err := json.Marshal(t.Values)
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(t.Schema, t.Table, t.Verb, string(outBytes), t.Retries)
	if err != nil {
		log.Fatal(err)
	}
}
