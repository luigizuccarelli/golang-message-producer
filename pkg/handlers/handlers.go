package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"gitea-cicd.apps.aws2-dev.ocp.14west.io/cicd/trackmate-message-producer/pkg/connectors"
	"gitea-cicd.apps.aws2-dev.ocp.14west.io/cicd/trackmate-message-producer/pkg/schema"
)

const (
	CONTENTTYPE     string = "Content-Type"
	APPLICATIONJSON string = "application/json"
)

// StreamHandler a http response and request for a message producer
func StreamHandler(w http.ResponseWriter, r *http.Request, conn connectors.Clients) {

	var response *schema.Response

	addHeaders(w, r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response = &schema.Response{StatusCode: "500", Status: "ERROR", Message: "Could not read body data " + err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		err = conn.SendMessageSync(body)
		if err != nil {
			response = &schema.Response{StatusCode: "500", Status: "ERROR", Message: "Could not send stream data " + err.Error()}
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			response = &schema.Response{StatusCode: "200", Status: "OK", Message: "Stream data sent successfully"}
		}
	}

	b, _ := json.MarshalIndent(response, "", "	")
	fmt.Fprintf(w, string(b))
}

// IsAlive a http response and request wrapper for health endpoint checks
// It takes a both response and request objects and returns void
func IsAlive(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{\"version\": \""+os.Getenv("VERSION")+"\"}")
}

// headers (with cors) utility
func addHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(CONTENTTYPE, APPLICATIONJSON)
	// use this for cors
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
