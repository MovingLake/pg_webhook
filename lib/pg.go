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
	"context"
	"log"

	"github.com/jackc/pgconn"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgtype"
)

func InitializeReplicationConnection(pgDns, outputPlugin, slotName string) (*pglogrepl.IdentifySystemResult, *pgconn.PgConn, error) {
	conn, err := pgconn.Connect(context.Background(), pgDns)
	if err != nil {
		log.Println("failed to connect to PostgreSQL server:", err)
		return nil, nil, err
	}

	result := conn.Exec(context.Background(), "DROP PUBLICATION IF EXISTS "+slotName+";")
	_, err = result.ReadAll()
	if err != nil {
		log.Println("drop publication if exists error", err)
		return nil, nil, err
	}

	result = conn.Exec(context.Background(), "CREATE PUBLICATION "+slotName+" FOR ALL TABLES;")
	_, err = result.ReadAll()
	if err != nil {
		log.Println("create publication error", err)
		return nil, nil, err
	}
	log.Println("create publication " + slotName)

	var pluginArguments []string
	if outputPlugin == "pgoutput" {
		pluginArguments = []string{"proto_version '1'", "publication_names '" + slotName + "'"}
	} else if outputPlugin == "wal2json" {
		pluginArguments = []string{"\"pretty-print\" 'true'"}
	}

	sysident, err := pglogrepl.IdentifySystem(context.Background(), conn)
	if err != nil {
		log.Println("IdentifySystem failed:", err)
		return nil, nil, err
	}
	log.Println("SystemID:", sysident.SystemID, "Timeline:", sysident.Timeline, "XLogPos:", sysident.XLogPos, "DBName:", sysident.DBName)

	_, err = pglogrepl.CreateReplicationSlot(context.Background(), conn, slotName, outputPlugin, pglogrepl.CreateReplicationSlotOptions{Temporary: true})
	if err != nil {
		log.Println("CreateReplicationSlot failed:", err)
		return nil, nil, err
	}
	log.Println("Created temporary replication slot:", slotName)
	err = pglogrepl.StartReplication(context.Background(), conn, slotName, sysident.XLogPos, pglogrepl.StartReplicationOptions{PluginArgs: pluginArguments})
	if err != nil {
		log.Println("StartReplication failed:", err)
		return nil, nil, err
	}
	log.Println("Logical replication started on slot", slotName)
	return &sysident, conn, nil
}

func GetValuesFromColumns(columns []*pglogrepl.TupleDataColumn, rel *pglogrepl.RelationMessage, info *pgtype.ConnInfo) map[string]interface{} {
	values := map[string]interface{}{}
	for idx, col := range columns {
		colName := rel.Columns[idx].Name
		switch col.DataType {
		case 'n': // null
			values[colName] = nil
		case 'u': // unchanged toast
			// This TOAST value was not changed. TOAST values are not stored in the tuple, and logical replication doesn't want to spend a disk read to fetch its value for you.
		case 't': //text
			val, err := decodeTextColumnData(info, col.Data, rel.Columns[idx].DataType)
			if err != nil {
				log.Fatalln("error decoding column data:", err)
			}
			values[colName] = val
		}
	}
	return values
}

func decodeTextColumnData(ci *pgtype.ConnInfo, data []byte, dataType uint32) (interface{}, error) {
	var decoder pgtype.TextDecoder
	if dt, ok := ci.DataTypeForOID(dataType); ok {
		decoder, ok = dt.Value.(pgtype.TextDecoder)
		if !ok {
			decoder = &pgtype.GenericText{}
		}
	} else {
		decoder = &pgtype.GenericText{}
	}
	if err := decoder.DecodeText(ci, data); err != nil {
		return nil, err
	}
	return decoder.(pgtype.Value).Get(), nil
}
