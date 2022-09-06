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

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"github.com/movinglake/pg_webhook/lib"
	"github.com/movinglake/pg_webhook/providers"
)

const (
	outputPlugin = "pgoutput"
)

var (
	pgDns                  = os.Getenv("PG_DNS")
	webhookServiceUrl      = os.Getenv("WEBHOOK_SERVICE_URL")
	slotName               = os.Getenv("REPLICATION_SLOT_NAME")
	maxConsumerConcurrency = 100
)

func main() {
	if pgDns == "" {
		log.Fatal("PG_DNS env variable is required")
	}
	if webhookServiceUrl == "" {
		log.Fatal("WEBHOOK_SERVICE_URL env variable is required")
	}
	if slotName == "" {
		slotName = "pg_webhook_slot"
	}
	sysident, conn, err := lib.InitializeReplicationConnection(pgDns, outputPlugin, slotName)
	defer conn.Close(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	clientXLogPos := sysident.XLogPos
	standbyMessageTimeout := time.Second * 10
	nextStandbyMessageDeadline := time.Now().Add(standbyMessageTimeout)
	relations := map[uint32]*pglogrepl.RelationMessage{}
	connInfo := pgtype.NewConnInfo()

	tasks := make(chan providers.Task)
	handler := &providers.SqliteHandler{}
	httpClient := &providers.HttpClientImpl{}
	handler.OpenDB()
	defer handler.Close()
	for i := 0; i < maxConsumerConcurrency; i++ {
		go lib.ProcessWebhookTasks(tasks, webhookServiceUrl, handler, httpClient)
	}

	for {
		if time.Now().After(nextStandbyMessageDeadline) {
			err = pglogrepl.SendStandbyStatusUpdate(context.Background(), conn, pglogrepl.StandbyStatusUpdate{WALWritePosition: clientXLogPos})
			if err != nil {
				log.Fatalln("SendStandbyStatusUpdate failed:", err)
			}
			log.Println("Sent Standby status message")
			nextStandbyMessageDeadline = time.Now().Add(standbyMessageTimeout)
		}

		ctx, cancel := context.WithDeadline(context.Background(), nextStandbyMessageDeadline)
		rawMsg, err := conn.ReceiveMessage(ctx)
		cancel()
		if err != nil {
			if pgconn.Timeout(err) {
				continue
			}
			log.Fatalln("ReceiveMessage failed:", err)
		}

		if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
			log.Printf("received Postgres WAL error: %+v", errMsg)
			return
		}

		msg, ok := rawMsg.(*pgproto3.CopyData)
		if !ok {
			log.Printf("Received unexpected message: %T\n", rawMsg)
			continue
		}

		switch msg.Data[0] {
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			pkm, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
			if err != nil {
				log.Fatalln("ParsePrimaryKeepaliveMessage failed:", err)
			}
			log.Println("Primary Keepalive Message =>", "ServerWALEnd:", pkm.ServerWALEnd, "ServerTime:", pkm.ServerTime, "ReplyRequested:", pkm.ReplyRequested)

			if pkm.ReplyRequested {
				nextStandbyMessageDeadline = time.Time{}
			}

		case pglogrepl.XLogDataByteID:
			xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				log.Fatalln("ParseXLogData failed:", err)
			}
			log.Println("XLogData =>", "WALStart", xld.WALStart, "ServerWALEnd", xld.ServerWALEnd, "ServerTime:", xld.ServerTime, "WALData", string(xld.WALData))
			logicalMsg, err := pglogrepl.Parse(xld.WALData)
			if err != nil {
				log.Fatalf("Parse logical replication message: %s", err)
			}
			log.Printf("Receive a logical replication message: %s", logicalMsg.Type())
			switch logicalMsg := logicalMsg.(type) {
			case *pglogrepl.RelationMessage:
				relations[logicalMsg.RelationID] = logicalMsg

			case *pglogrepl.BeginMessage:
				// Indicates the beginning of a group of changes in a transaction. This is only sent for committed transactions. You won't get any events from rolled back transactions.

			case *pglogrepl.CommitMessage:

			case *pglogrepl.InsertMessage:
				rel, ok := relations[logicalMsg.RelationID]
				if !ok {
					log.Fatalf("unknown relation ID %d", logicalMsg.RelationID)
				}
				values := lib.GetValuesFromColumns(logicalMsg.Tuple.Columns, rel, connInfo)
				log.Printf("INSERT INTO %s.%s: %v", rel.Namespace, rel.RelationName, values)

				tasks <- providers.Task{
					Ctx:    ctx,
					Schema: rel.Namespace,
					Table:  rel.RelationName,
					Verb:   "INSERT",
					Values: values,
				}

			case *pglogrepl.UpdateMessage:
				rel, ok := relations[logicalMsg.RelationID]
				if !ok {
					log.Fatalf("unknown relation ID %d", logicalMsg.RelationID)
				}
				values := lib.GetValuesFromColumns(logicalMsg.NewTuple.Columns, rel, connInfo)
				log.Printf("UPDATE %s.%s: %v", rel.Namespace, rel.RelationName, values)
				tasks <- providers.Task{
					Ctx:    ctx,
					Schema: rel.Namespace,
					Table:  rel.RelationName,
					Verb:   "UPDATE",
					Values: values,
				}
			case *pglogrepl.DeleteMessage:
				rel, ok := relations[logicalMsg.RelationID]
				if !ok {
					log.Fatalf("unknown relation ID %d", logicalMsg.RelationID)
				}
				values := lib.GetValuesFromColumns(logicalMsg.OldTuple.Columns, rel, connInfo)
				log.Printf("DELETE %s.%s: %v", rel.Namespace, rel.RelationName, values)
				tasks <- providers.Task{
					Ctx:    ctx,
					Schema: rel.Namespace,
					Table:  rel.RelationName,
					Verb:   "DELETE",
					Values: values,
				}
			case *pglogrepl.TruncateMessage:
				// ...

			case *pglogrepl.TypeMessage:
			case *pglogrepl.OriginMessage:
			default:
				log.Printf("Unknown message type in pgoutput stream: %T", logicalMsg)
			}

			clientXLogPos = xld.WALStart + pglogrepl.LSN(len(xld.WALData))
		}
	}
}
