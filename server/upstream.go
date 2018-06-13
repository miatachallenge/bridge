package main

import (
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/miatachallenge/bridge/server/debouncer"
	"github.com/pkg/errors"
)

var (
	triggerDebouncer = debouncer.New(time.Second)
	debouncedTrigger = make(chan struct{}, 10240)
)

func startUpstream() {
	go func() {
		for {
			<-triggerDebouncer.Output
			debouncedTrigger <- struct{}{}
			// Try to sync
		}
	}()

	go func() {
		for {
			if err := func() error {
				defer time.Sleep(time.Second)

				conn, _, err := websocket.DefaultDialer.Dial(*upstream, nil)
				if err != nil {
					return errors.Wrap(err, "unable to dial")
				}
				defer conn.Close()

				stateLock.Lock()
				state.Upstream.Connected = true
				stateLock.Unlock()

				defer func() {
					stateLock.Lock()
					state.Upstream.Connected = false
					stateLock.Unlock()
				}()

				if err := conn.WriteJSON(map[string]interface{}{
					"type": "auth",
					"key":  *upstreamKey,
				}); err != nil {
					return errors.Wrap(err, "unable to write auth")
				}

				kill := make(chan struct{})
				defer close(kill)

				go func() {
					defer func() {
						kill <- struct{}{}
					}()

					for {
						var msg struct {
							Type    string `json:"type"`
							Current int    `json:"current"`
						}
						if err := conn.ReadJSON(&msg); err != nil {
							log.Printf("websockets: unable to read json - %s", err)

							if _, ok := err.(*websocket.CloseError); ok || strings.Contains(err.Error(), "reset by peer") {
								return
							}

							continue
						}

						if msg.Type == "current" {
							stateLock.Lock()
							state.Upstream.Count = msg.Current
							shouldUpdate := state.Database.Count > state.Upstream.Count
							stateLock.Unlock()

							if shouldUpdate {
								triggerDebouncer.Trigger()
							}
						}
					}
				}()

				for {
					select {
					case <-debouncedTrigger:
						stateLock.RLock()
						var (
							databaseCount = state.Database.Count
							upstreamCount = state.Upstream.Count
						)
						stateLock.RUnlock()

						log.Printf("Going to sync from %d to %d", upstreamCount, databaseCount)

						limit := databaseCount - upstreamCount
						if limit > 100 {
							limit = 100
						}

						events := []*rfidEvent{}
						if err := db.Select(
							&events,
							"SELECT * FROM records LIMIT ?, ?",
							upstreamCount, limit,
						); err != nil {
							return errors.Wrap(err, "unable to get records")
						}

						if err := conn.WriteJSON(map[string]interface{}{
							"type":   "sync",
							"events": events,
						}); err != nil {
							return errors.Wrap(err, "unable to upload the records")
						}
					case <-kill:
						return nil
					}
				}

				return nil
			}(); err != nil {
				log.Printf("upstream: %s", err)
			}
		}
	}()
}
