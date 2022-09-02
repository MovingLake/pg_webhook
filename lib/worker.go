/*
 * This file is part of the pg_webhook distribution (https://github.com/pg_webhook).
 * Copyright (c) 2022 MovingLake Intermediate Holdings LLC.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package lib

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	maxRetries = 2 // Around 22 hours.
)

type Task struct {
	Ctx     context.Context
	Schema  string
	Table   string
	Verb    string
	Values  map[string]interface{}
	Retries int
}

func ProcessWebhookTasks(tasks chan Task, webhookServiceUrl string, db *sql.DB) {
	for {
		t := <-tasks

		if t.Retries > maxRetries {
			log.Println("max retries reached, sending task to sqlite")
			InsertTask(db, t)
			continue
		}

		t.Values["verb"] = t.Verb
		t.Values["schema"] = t.Schema
		t.Values["table"] = t.Table

		outBytes, err := json.Marshal(t.Values)
		if err != nil {
			log.Printf("failed to marshal values: %v", err)
			continue
		}

		resp, err := http.Post(webhookServiceUrl, "application/json", bytes.NewBuffer(outBytes))

		if err != nil {
			log.Printf("failed to send to webhook service: %v", err)
			t.Retries += 1
			time.Sleep(time.Duration(t.Retries) * time.Second)
			tasks <- t
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("failed to send to webhook service: %v", resp)
			t.Retries += 1
			time.Sleep(time.Duration(t.Retries) * time.Second)
			tasks <- t
			continue
		}
	}
}

func OpenDB() *sql.DB {
	db, err := sql.Open("sqlite3", "db.sqlite")

	if err != nil {
		log.Fatal(err)
		return nil
	}
	return db
}

func InsertTask(db *sql.DB, t Task) {
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS tasks (schema TEXT, table_name TEXT, verb TEXT, values_table TEXT, retries INT)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}
	stmt.Close()

	stmt, err = db.Prepare("INSERT INTO tasks (schema, table_name, verb, values_table, retries) VALUES (?, ?, ?, ?, ?)")
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
