package main

import (
	"log"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sqlx.DB
)

func startDatabase() {
	var err error
	db, err = sqlx.Open("sqlite3", *database)
	if err != nil {
		panic(err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS records (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		tag_id    TEXT,
		timestamp NUMBER,
		antenna   NUMBER
	)`); err != nil {
		panic(err)
	}

	rows, err := db.Queryx("SELECT * FROM records ORDER BY timestamp DESC LIMIT 20")
	if err != nil {
		panic(err)
	}

	events := []*rfidEvent{}
	for rows.Next() {
		event := &rfidEvent{}
		if err := rows.StructScan(event); err != nil {
			panic(err)
		}

		events = append(events, event)
	}

	var count int
	if err := db.Get(&count, "SELECT COUNT(*) FROM records"); err != nil {
		panic(err)
	}

	stateLock.Lock()
	state.Readouts = events
	state.Database.Count = count
	stateLock.Unlock()

	go func() {
		stateLock.Lock()
		state.Database.Running = true
		stateLock.Unlock()

		for {
			readout := <-readouts

			log.Printf("%+v", readout)

			if _, err := db.Exec(
				"INSERT INTO records(tag_id, timestamp, antenna) VALUES(?, ?, ?)",
				readout.TagID, readout.Timestamp, readout.Antenna,
			); err != nil {
				log.Printf("database: failed to insert - %s", err)
				continue
			}

			var count int
			if err := db.Get(&count, "SELECT COUNT(*) FROM records"); err != nil {
				log.Printf("database: failed to to count records - %s", err)
				continue
			}

			stateLock.Lock()
			state.Database.Count = count
			state.Readouts = append(state.Readouts, readout)
			if len(state.Readouts) > rfidMaxReadouts {
				state.Readouts = state.Readouts[len(state.Readouts)-rfidMaxReadouts:]
			}
			stateLock.Unlock()

			triggerDebouncer.Trigger()
		}
	}()
}
