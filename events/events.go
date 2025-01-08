package events

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
)

var MessageChan = make(chan string)

func EventStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Encoding", "identity")
	w.Header().Set("X-Accel-Buffering", "no")

	ctx := r.Context()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add the directory to be watched
	err = watcher.Add("templates")
	if err != nil {
		log.Fatal(err)
	}
	// Add the directory to be watched
	err = watcher.Add("static")
	if err != nil {
		log.Fatal(err)
	}
	// Send initial "connected" message
	msg := "event: connected\ndata: connected\n\n"
	fmt.Printf("writing connected message\n")
	_, writeErr := fmt.Fprint(w, msg)
	if writeErr != nil {
		// Client has likely disconnected, log error and exit
		log.Printf("Error writing to client: %v", writeErr)
		return
	}
	w.(http.Flusher).Flush()

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				msg := "event: reload\ndata: reload\n\n"
				log.Printf("writing reload message")
				_, writeErr := fmt.Fprint(w, msg)
				if writeErr != nil {
					// Client has likely disconnected, log error and exit
					log.Printf("Error writing to client: %v", writeErr)
					return
				}
				w.(http.Flusher).Flush()
			}
		case t := <-ticker.C:
			msg := fmt.Sprintf("event: time\ndata: The server time is: %v\n\n", t)
			log.Printf("writing message")
			_, writeErr := fmt.Fprint(w, msg)
			if writeErr != nil {
				// Client has likely disconnected, log error and exit
				log.Printf("Error writing to client: %v", writeErr)
				return
			}
			w.(http.Flusher).Flush()
		case msg := <-MessageChan:
			_msg := fmt.Sprintf("event: entity \ndata: %v\n\n", msg)
			log.Printf("writing message")
			_, writeErr := fmt.Fprint(w, _msg)
			if writeErr != nil {
				log.Printf("Error writing to client: %v", writeErr)
				return
			}
			w.(http.Flusher).Flush()
		case err := <-watcher.Errors:
			log.Printf("Watcher error: %v", err)
		case <-ctx.Done():
			log.Println("Client disconnected, stopping eventStream")
			return
		}
	}
}

func EmitEvent(event string, entity interface{}) {
	if entity != nil {
		entityData, err := json.Marshal(entity)
		if err != nil {
			log.Printf("Error marshaling entity: %v", err)
			return
		}
		MessageChan <- fmt.Sprintf("{\"%s\": %s}", event, string(entityData))
	} else {
		MessageChan <- event
	}
}
