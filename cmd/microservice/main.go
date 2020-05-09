package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gitea-cicd.apps.aws2-dev.ocp.14west.io/cicd/trackmate-message-producer/pkg/connectors"
	"gitea-cicd.apps.aws2-dev.ocp.14west.io/cicd/trackmate-message-producer/pkg/handlers"
	"gitea-cicd.apps.aws2-dev.ocp.14west.io/cicd/trackmate-message-producer/pkg/validator"
	"github.com/gorilla/mux"
	"github.com/microlib/simple"
)

// startHttpServer a utility function that sets the routes, handlers and starts the http server
func startHttpServer(conn connectors.Clients) *http.Server {

	// set the server props
	srv := &http.Server{Addr: ":" + os.Getenv("SERVER_PORT")}

	// set the router and endpoints
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/streamdata", func(w http.ResponseWriter, req *http.Request) {
		handlers.StreamHandler(w, req, conn)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/v2/sys/info/isalive", handlers.IsAlive).Methods("GET", "OPTIONS")

	http.Handle("/", r)

	// start our server (concurrent)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			conn.Error("Httpserver: ListenAndServe() error: " + err.Error())
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

	var logger *simple.Logger

	if os.Getenv("LOG_LEVEL") == "" {
		logger = &simple.Logger{Level: "info"}
	} else {
		logger = &simple.Logger{Level: os.Getenv("LOG_LEVEL")}
	}
	err := validator.ValidateEnvars(logger)
	if err != nil {
		os.Exit(-1)
	}

	// setup our client connectors (message producer)
	conn := connectors.NewClientConnectors(logger)

	// call the start server function
	srv := startHttpServer(conn)
	logger.Info("Starting server on port " + os.Getenv("SERVER_PORT"))
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

	conn.Close()

	if err := srv.Shutdown(nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to shut down server cleanly %v", err))
	}

	logger.Info("Server shutdown successfully")
	os.Exit(code)

}
