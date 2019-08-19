package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/microlib/simple"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	logger     simple.Logger
	connectors Clients
	counter    int = 0
)

type Clients interface {
	SendMessageSync(body []byte) error
}

type FakeProducer struct {
}

type Connectors struct {
	producer FakeProducer
	Name     string
	Http     *http.Client
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewHttpTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func NewTestClients(data string, code int) Clients {

	// read the config
	//config, _ = Init("config.json")
	logger.Level = "debug"

	// we first load the json payload to simulate a call to middleware
	// for now just ignore failures.
	file, _ := ioutil.ReadFile(data)
	logger.Trace(fmt.Sprintf("File %s with data %s", data, string(file)))
	httpclient := NewHttpTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: code,
			// Send response to be tested

			Body: ioutil.NopCloser(bytes.NewBufferString(string(file))),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	p := FakeProducer{}
	conns := &Connectors{producer: p, Name: "test", Http: httpclient}
	return conns
}

func (r *Connectors) SendMessageSync(b []byte) error {
	// We are not setting a message key, which means that all messages will
	// be distributed randomly over the different partitions.
	logger.Info(fmt.Sprintf("Byte array %s", string(b)))
	if string(b) == "{\"error\"}" {
		return errors.New("Error byte buffer")
	}
	return nil
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestAll(t *testing.T) {

	var req *http.Request
	var response Response

	// create anonymous struct
	tests := []struct {
		Name     string
		Method   string
		Url      string
		Payload  string
		Handler  string
		FileName string
		Want     int
		ErrorMsg string
	}{
		{
			"Testing options method : should pass",
			"OPTIONS",
			"/sys/info/isalive",
			"",
			"IsAlive",
			"tests/payload-example.json",
			200,
			"Handler %s returned - got (%v) wanted (%v)",
		},
		{
			"Testing isalive endpoint : should pass",
			"GET",
			"/sys/info/isalive",
			"",
			"IsAlive",
			"tests/payload-example.json",
			200,
			"Handler %s returned - got (%v) wanted (%v)",
		},
		{
			"Testing streamdata endpoint : should pass",
			"POST",
			"/streamdata",
			"{\"event\":\"click\", \"target\":\"sort-stock\"}",
			"Streamer",
			"tests/payload-example.json",
			200,
			"Handler %s returned - got (%v) wanted (%v)",
		},
		{
			"Testing streamdata endpoint : should fail",
			"POST",
			"/streamdata",
			"",
			"Streamer",
			"tests/payload-example.json",
			500,
			"Handler %s returned - got (%v) wanted (%v)",
		},
		{
			"Testing streamdata endpoint : should fail",
			"POST",
			"/streamdata",
			"{\"error\"}",
			"Streamer",
			"tests/payload-example.json",
			500,
			"Handler %s returned - got (%v) wanted (%v)",
		},
	}

	for _, tt := range tests {
		logger.Info(fmt.Sprintf("Executing test : %s \n", tt.Name))
		if tt.Payload == "" {
			req, _ = http.NewRequest(tt.Method, tt.Url, errReader(0))
		} else {
			req, _ = http.NewRequest(tt.Method, tt.Url, bytes.NewBuffer([]byte(tt.Payload)))
		}

		connectors = NewTestClients(tt.FileName, tt.Want)

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		switch tt.Handler {
		case "IsAlive":
			req.Header.Set("API-KEY", "test1234")
			handler := http.HandlerFunc(IsAlive)
			handler.ServeHTTP(rr, req)
		case "Streamer":
			handler := http.HandlerFunc(Streamer)
			handler.ServeHTTP(rr, req)
		}
		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		body, e := ioutil.ReadAll(rr.Body)
		if e != nil {
			t.Fatalf(fmt.Sprintf(tt.ErrorMsg, tt.Handler, "nil", "error"))
		}
		// ignore errors here
		json.Unmarshal(body, &response)
		if rr.Code != tt.Want {
			t.Fatalf(fmt.Sprintf(tt.ErrorMsg, tt.Handler, "nil", "error"))
		}
	}
}
