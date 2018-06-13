package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	rfidConsolePort     = 50007
	rfidEventPort       = 50008
	rfidTimestampFormat = "2006-01-02T15:04:05.999"
	rfidMaxReadouts     = 10
)

type rfidEvent struct {
	ID        int64  `json:"id" db:"id"`
	Name      string `json:"-" db:"name"`
	TagID     string `json:"tag_id" db:"tag_id"`
	Timestamp int64  `json:"timestamp" db:"timestamp"`
	Antenna   int    `json:"antenna" db:"antenna"`
}

var (
	readouts = make(chan *rfidEvent)
)

func connectRFID() error {
	rawConsoleConn, err := net.DialTimeout(
		"tcp", *modem+":"+strconv.Itoa(rfidConsolePort),
		time.Second*5,
	)
	if err != nil {
		return errors.Wrap(err, "unable to establish the console conn")
	}
	defer rawConsoleConn.Close()

	consoleReader := bufio.NewReader(rawConsoleConn)

	rawEventsConn, err := net.DialTimeout(
		"tcp", *modem+":"+strconv.Itoa(rfidEventPort),
		time.Second*5,
	)
	if err != nil {
		return errors.Wrap(err, "unable to establish the events conn")
	}
	defer rawEventsConn.Close()

	eventsReader := bufio.NewReader(rawEventsConn)

	stateLock.Lock()
	state.Modem.Running = true
	stateLock.Unlock()

	defer func() {
		stateLock.Lock()
		state.Modem.Running = false
		stateLock.Unlock()
	}()

	go func() {
		for {
			line, _, err := consoleReader.ReadLine()
			if err != nil {
				return
			}
			if len(line) == 0 {
				continue
			}

			log.Printf("rfid console: %s", line)
		}
	}()

	eventsID, _, err := eventsReader.ReadLine()
	if err != nil {
		return errors.Wrap(err, "unable to get the events connection ID")
	}

	// 4th part is the number
	eventsIDParts := strings.SplitN(string(eventsID), " ", 4)
	log.Printf("Acquired a TCP events channel with id %s. Registering it now.", eventsIDParts[3])

	// Register a subscription
	rawConsoleConn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	if _, err := fmt.Fprintf(rawConsoleConn,
		"reader.events.register(id=%s,name=event.tag.arrive)\n", eventsIDParts[3],
	); err != nil {
		return errors.Wrap(err, "unable to create a new subscription")
	}
	if _, err := fmt.Fprintf(rawConsoleConn,
		"setup.operating_mode=active",
	); err != nil {
		return errors.Wrap(err, "unable to enable reading mode")
	}

	for {
		rawEventsConn.SetReadDeadline(time.Now().Add(time.Second * 300))
		line, _, err := eventsReader.ReadLine()
		if err != nil {
			return errors.Wrap(err, "unable to read a line")
		}
		if len(line) == 0 {
			continue
		}

		// name a=b, c=d, e=f
		lineSplit := strings.SplitN(string(line), " ", 2)
		paramPairs := strings.Split(lineSplit[1], ", ")

		event := &rfidEvent{
			Name: lineSplit[0],
		}
		for _, pair := range paramPairs {
			kv := strings.SplitN(pair, "=", 2)
			switch kv[0] {
			case "tag_id":
				event.TagID = kv[1]
			case "first":
				timestamp, err := time.Parse(rfidTimestampFormat, kv[1])
				if err != nil {
					log.Printf("rfid events: unable to parse timestamp %s - %s", kv[1], err)
					continue
				}

				event.Timestamp = timestamp.UnixNano() / 1000000
			case "antenna":
				antenna, err := strconv.Atoi(kv[1])
				if err != nil {
					log.Printf("rfid events: unable to parse antenna %s - %s", kv[1], err)
					continue
				}

				event.Antenna = antenna
			}
		}

		readouts <- event
	}

	return nil
}

func startRFID() {
	go func() {
		for {
			if err := connectRFID(); err != nil {
				log.Printf("rfid: %s", err)
			}

			time.Sleep(time.Second)
		}
	}()
}
