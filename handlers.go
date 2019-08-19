package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	CONTENTTYPE     string = "Content-Type"
	APPLICATIONJSON string = "application/json"
)

// Streamer a http response and request for a message producer
// @param - http.ResponseWriter (used to send back response)
// @param - http.Request object
func Streamer(w http.ResponseWriter, r *http.Request) {

	var response Response

	addHeaders(w, r)
	handleOptions(w, r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response = Response{StatusCode: "500", Status: "ERROR", Message: "Could not read body data " + err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		err = connectors.SendMessageSync(body)
		if err != nil {
			response = Response{StatusCode: "500", Status: "ERROR", Message: "Could not send stream data " + err.Error()}
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			response = Response{StatusCode: "200", Status: "OK", Message: "Stream data sent successfully"}
		}
	}

	b, _ := json.MarshalIndent(response, "", "	")
	fmt.Fprintf(w, string(b))
}

// IsAlive a http response and request wrapper for health endpoint checks
// It takes a both response and request objects and returns void
func IsAlive(w http.ResponseWriter, r *http.Request) {
	addHeaders(w, r)
	handleOptions(w, r)
	logger.Trace(fmt.Sprintf("used to mask cc %v", r))
	fmt.Fprintf(w, "{\"version\": \""+os.Getenv("VERSION")+"\"}")
}

// headers (with cors) utility
func addHeaders(w http.ResponseWriter, r *http.Request) {
	var request []string
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	logger.Debug(fmt.Sprintf("Headers : %s", request))

	w.Header().Set(CONTENTTYPE, APPLICATIONJSON)
	// use this for cors
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// simple options handler
func handleOptions(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "")
	}
	return
}
