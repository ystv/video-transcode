package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/event"
	"github.com/ystv/video-transcode/manager"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		log.Fatalf("failed to connect to mq: %+v", err)
	}

	emitter, err := event.NewProducer(conn)
	if err != nil {
		log.Fatalf("failed to start producer: %+v", err)
	}

	m := manager.New(emitter)

	r := mux.NewRouter()
	mount(r, "/", m.Router())

	log.Printf("listening on :7071")
	log.Fatal(http.ListenAndServe(":7071", r))
}

// mount another mux router ontop of another
func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
}
