package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/event"
)

var emitter event.Producer

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		panic(err)
	}

	emitter, err = event.NewProducer(conn)
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandle)
	r.HandleFunc("/new_vod", NewVODHandle)
	r.HandleFunc("/new_live", NewLiveHandle)
	log.Printf("listening on :7071")
	log.Fatal(http.ListenAndServe(":7071", r))
}
