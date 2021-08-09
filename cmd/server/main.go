package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"github.com/ystv/video-transcode/event"
	"github.com/ystv/video-transcode/manager"
)

// Config represents VT's configuration
type Config struct {
	AMQPEndpoint string
	HTTPUser     string
	HTTPPass     string
	HTTPPort     string
}

var conf Config

func main() {
	godotenv.Load(".env.local")
	godotenv.Load(".env")
	conf.AMQPEndpoint = os.Getenv("VT_AMQP_ENDPOINT")
	conf.HTTPPort = os.Getenv("VT_HTTP_PORT")
	conf.HTTPUser = os.Getenv("VT_HTTP_USER")
	conf.HTTPPass = os.Getenv("VT_HTTP_PASS")

	if conf.HTTPPort == "" {
		conf.HTTPPort = "7071"
	}

	conn, err := amqp.Dial(conf.AMQPEndpoint)
	if err != nil {
		log.Fatalf("failed to connect to mq: %+v", err)
	}

	emitter, err := event.NewEventer(conn)
	if err != nil {
		log.Fatalf("failed to start eventer: %+v", err)
	}

	m := manager.New(emitter, conf.HTTPUser, conf.HTTPPass)

	r := mux.NewRouter()
	mount(r, "/", m.Router())

	log.Printf("listening on :%s", conf.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+conf.HTTPPort, r))
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
