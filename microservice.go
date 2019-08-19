package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/microlib/simple"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	logger     simple.Logger
	connectors Clients
)

// startHttpServer a utility function that sets the routes, handlers and starts the http server
func startHttpServer(cd ConnectionData) *http.Server {

	// set the server props
	srv := &http.Server{Addr: ":" + cd.Port}

	// set the router and endpoints
	r := mux.NewRouter()
	r.HandleFunc("/streamdata", Streamer).Methods("POST")
	r.HandleFunc("/sys/info/isalive", IsAlive).Methods("GET", "OPTIONS")
	http.Handle("/", r)

	// start our server (concurrent)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("Httpserver: ListenAndServe() error: " + err.Error())
			os.Exit(0)
		}
	}()

	// return our srv object
	return srv
}

// The main function reads the config file, parsers and validates it and calls our start server function
// using a go channel to intercept sig calls from the os
// A simple curl to test the payload endpoint
// curl -H "Accept: application/json"  -H "Content-Type: application/json" -X PUT -d @sparkpost-webhook-payload.json http://sparkpost-spring-producer-microservice-sparkpost-poc.apps.poc.okd.14west.io/webhook
func main() {

	ValidateEnvars()

	connectionData := ConnectionData{
		Name:    "RealConnector",
		Port:    os.Getenv("SERVER_PORT"),
		Brokers: os.Getenv("KAFKA_BROKERS"),
		Topic:   os.Getenv("TOPIC"),
	}

	// set the logger level
	logger.Level = os.Getenv("LOG_LEVEL")

	// setup our client connectors (message producer)
	connectors = NewClientConnectors(connectionData)

	// call the start server function
	srv := startHttpServer(connectionData)
	logger.Info("Starting server on port " + connectionData.Port)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// set up a channel
	exit_chan := make(chan int)

	// listen for signals
	go func() {
		for {
			s := <-c
			switch s {
			case syscall.SIGHUP:
				exit_chan <- 0
			case syscall.SIGINT:
				exit_chan <- 0
			case syscall.SIGTERM:
				exit_chan <- 0
			case syscall.SIGQUIT:
				exit_chan <- 0
			default:
				exit_chan <- 1
			}
		}
	}()

	// cleanup and shutdown
	code := <-exit_chan

	connectors.Close()

	if err := srv.Shutdown(nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to shut down server cleanly", err))

	}

	logger.Info("Server shutdown successfully")
	os.Exit(code)

}

func checkEnvar(name string, required bool) {
	if os.Getenv(name) == "" {
		if required {
			logger.Error(fmt.Sprintf("%s envar is mandatory please set it", name))
			os.Exit(-1)
		} else {
			logger.Error(fmt.Sprintf("%s envar is empty please set it", name))
		}
	}
}

func ValidateEnvars() {
	checkEnvar("LOG_LEVEL", false)
	checkEnvar("SERVER_PORT", false)
	checkEnvar("KAFKA_BROKERS", true)
	checkEnvar("TOPIC", true)
}
