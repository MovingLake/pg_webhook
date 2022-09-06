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
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/movinglake/pg_webhook/providers"
)

const (
	MaxRetries = 400 // Around 22 hours.
)

func ProcessWebhookTasks(tasks chan providers.Task, webhookServiceUrl string, handler providers.DbHandler, httpClient providers.HttpClient) {
	for {
		t := <-tasks

		if t.Retries > MaxRetries {
			log.Println("max retries reached, sending task to sqlite")
			handler.InsertTask(t)
			continue
		}

		t.Values["_mlake_verb"] = t.Verb
		t.Values["_mlake_schema"] = t.Schema
		t.Values["_mlake_table"] = t.Table

		outBytes, err := json.Marshal(t.Values)
		if err != nil {
			log.Printf("failed to marshal values: %v", err)
			continue
		}

		resp, err := httpClient.Post(webhookServiceUrl, "application/json", bytes.NewBuffer(outBytes))

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
