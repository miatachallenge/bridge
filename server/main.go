package main

import (
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	bind        = flag.String("bind", ":8080", "bind address")
	modem       = flag.String("modem", "192.168.1.151", "modem ip")
	database    = flag.String("database", "./database.sqlite3", "db path")
	upstream    = flag.String("upstream", "ws://mc.zduniak.net:8090/sync", "upstream")
	upstreamKey = flag.String("upstream_key", "", "upstream key")
)

type stateStruct struct {
	Modem struct {
		Address string `json:"address"`
		Ping    int64  `json:"ping"`
		Running bool   `json:"running"`
	} `json:"modem"`
	Upstream struct {
		Address   string `json:"address"`
		Connected bool   `json:"connected"`
		Count     int    `json:"count"`
	} `json:"upstream"`
	Database struct {
		Count   int  `json:"count"`
		Running bool `json:"running"`
	} `json:"database"`
	Readouts []*rfidEvent `json:"readouts"`
}

var (
	stateLock sync.RWMutex
	state     = stateStruct{
		Readouts: []*rfidEvent{},
	}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	flag.Parse()

	// Start the modem pinger
	state.Modem.Address = *modem
	state.Modem.Ping = -1
	startPinger()

	// Start the RFID client
	startRFID()

	// And the consumer
	startDatabase()

	// Start the upstream sync engine
	state.Upstream.Address = *upstream
	startUpstream()

	// WS handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		stateLock.RLock()
		err = conn.WriteJSON(state)
		stateLock.RUnlock()
		if err != nil {
			log.Printf("WS died: %s", err)
			return
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

	loop:
		for {
			select {
			case <-ticker.C:
				stateLock.RLock()
				err := conn.WriteJSON(state)
				stateLock.RUnlock()
				if err != nil {
					log.Printf("WS died: %s", err)
					break loop
				}
			}
		}
	})

	http.Handle("/", http.FileServer(http.Dir("frontend")))
	http.ListenAndServe(*bind, nil)
}
