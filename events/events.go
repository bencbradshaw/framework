package events

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var MessageChan = make(chan string)
var EventChan = make(chan string)

func EventStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Encoding", "iddata")
	w.Header().Set("X-Accel-Buffering", "no")
	ctx := r.Context()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
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
		case t := <-ticker.C:
			msg := fmt.Sprintf("event: time\ndata: The server time is: %v\n\n", t)
			log.Printf("writing message")
			_, writeErr := fmt.Fprint(w, msg)
			if writeErr != nil {
				log.Printf("Error writing to client: %v", writeErr)
				return
			}
			w.(http.Flusher).Flush()
		case msg := <-MessageChan:
			_msg := fmt.Sprintf("event: data \ndata: %v\n\n", msg)
			log.Printf("writing message")
			_, writeErr := fmt.Fprint(w, _msg)
			if writeErr != nil {
				log.Printf("Error writing to client: %v", writeErr)
				return
			}
			w.(http.Flusher).Flush()
		case eventMsg := <-EventChan:
			log.Printf("writing event message")
			_, writeErr := fmt.Fprint(w, eventMsg)
			if writeErr != nil {
				log.Printf("Error writing to client: %v", writeErr)
				return
			}
			w.(http.Flusher).Flush()
		case <-ctx.Done():
			log.Println("Client disconnected, stopping eventStream")
			return
		}
	}
}

func EmitEvent(event string, data interface{}) {
	if data != nil {
		dataData, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshaling data: %v", err)
			return
		}
		EventChan <- fmt.Sprintf("event: %s \ndata: %v\n\n", event, string(dataData))
	} else {
		EventChan <- fmt.Sprintf("event: %s \n", event)
	}
}
