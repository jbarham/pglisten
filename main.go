package main

// Based on https://pkg.go.dev/github.com/lib/pq/example/listen

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lib/pq"
)

func waitForNotification(l *pq.Listener) {
	for {
		select {
		case note := <-l.Notify:
			name := note.Extra
			if name == "" {
				name = "world"
			}
			log.Printf("Hello, %s!", name)
		case <-time.After(90 * time.Second):
			go l.Ping()
			// Check if there's more work available, just in case it takes
			// a while for the Listener to notice connection loss and
			// reconnect.
			log.Println("received no work for 90 seconds, checking for new work")
		}
	}
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("couldn't open DB: %s", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("couldn't ping DB: %s", err)
	}
	defer db.Close()

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Println(err.Error())
		}
	}

	minReconn := 10 * time.Second
	maxReconn := time.Minute
	listener := pq.NewListener(dbURL, minReconn, maxReconn, reportProblem)
	if err = listener.Listen("hello"); err != nil {
		log.Fatalf("couldn't start listener: %s", err)
	}

	go waitForNotification(listener)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	log.Print("Waiting for notifications, press Ctrl+C to exit...")
	<-done
}
