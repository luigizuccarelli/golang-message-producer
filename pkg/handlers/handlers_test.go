package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"lmzsoftware.com/lzuccarelli/golang-message-producer/pkg/connectors"
	"lmzsoftware.com/lzuccarelli/golang-message-producer/pkg/schema"
	"github.com/microlib/simple"
)

type FakeProducer struct {
}

type Connectors struct {
	Producer FakeProducer
	Http     *http.Client
	Logger   *simple.Logger
	Name     string
}

func (conn *Connectors) Close() {
}

func (conn *Connectors) Error(msg string, val ...interface{}) {
	conn.Logger.Error(fmt.Sprintf(msg, val...))
}

func (conn *Connectors) Info(msg string, val ...interface{}) {
	conn.Logger.Info(fmt.Sprintf(msg, val...))
}

func (conn *Connectors) Debug(msg string, val ...interface{}) {
	conn.Logger.Debug(fmt.Sprintf(msg, val...))
}

func (conn *Connectors) Trace(msg string, val ...interface{}) {
	conn.Logger.Trace(fmt.Sprintf(msg, val...))
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

func NewTestClient(data string, code int) connectors.Clients {

	logger := &simple.Logger{Level: "debug"}

	file, _ := ioutil.ReadFile(data)
	logger.Trace(fmt.Sprintf("File %s with data %s", data, string(file)))
	httpclient := NewHttpTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: code,
			// Send response to be tested

			Body: ioutil.NopCloser(bytes.NewBuffer(file)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	p := FakeProducer{}
	conns := &Connectors{Producer: p, Logger: logger, Name: "Test", Http: httpclient}
	return conns
}

func (conn *Connectors) SendMessageSync(b []byte) error {
	// We are not setting a message key, which means that all messages will
	// be distributed randomly over the different partitions.
	conn.Debug(fmt.Sprintf("Byte array %s", string(b)))
	if string(b) == "{\"error\"}" {
		return errors.New("Error byte buffer")
	}
	return nil
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("Test error")
}

func TestAll(t *testing.T) {

	var req *http.Request
	var response *schema.Response

	t.Run("Testing endpoint /sys/info/isalive : should pass", func(t *testing.T) {
		req, _ = http.NewRequest("GET", "/api/v2/sys/info/isalive", nil)
		rr := httptest.NewRecorder()
		req.Header.Set("API-KEY", "test1234")
		handler := http.HandlerFunc(IsAlive)
		handler.ServeHTTP(rr, req)
		body, e := ioutil.ReadAll(rr.Body)
		if e != nil {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "isAlive", e, nil))
		}
		// ignore errors here
		err := json.Unmarshal(body, &response)
		if rr.Code != 200 {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "isAlive", err, nil))
		}
	})

	t.Run("Testing endpoint OPTIONS /api/v1/streamdata : should pass", func(t *testing.T) {
		file, _ := ioutil.ReadFile("../../tests/segmentio.json")
		req, _ = http.NewRequest("OPTIONS", "/api/v1/streamdata", bytes.NewBuffer(file))
		rr := httptest.NewRecorder()
		conn := NewTestClient("../../tests/response.json", 200)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			StreamHandler(w, r, conn)
		})
		handler.ServeHTTP(rr, req)
		body, e := ioutil.ReadAll(rr.Body)
		if e != nil {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "StreamHandler", e, nil))
		}
		// ignore errors here
		err := json.Unmarshal(body, &response)
		conn.Debug(fmt.Sprintf("Response : %v", response))
		if rr.Code != 200 {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "StreamHandler", err, nil))
		}
	})

	t.Run("Testing endpoint POST /api/v1/streamdata : should fail", func(t *testing.T) {
		req, _ = http.NewRequest("POST", "/api/v1/streamdata", errReader(0))
		rr := httptest.NewRecorder()
		conn := NewTestClient("../../tests/response.json", 500)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			StreamHandler(w, r, conn)
		})
		handler.ServeHTTP(rr, req)
		body, e := ioutil.ReadAll(rr.Body)
		if e != nil {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "StreamHandler", e, nil))
		}
		// ignore errors here
		json.Unmarshal(body, &response)
		conn.Debug(fmt.Sprintf("Response : %v", response))
		if rr.Code != 500 {
			t.Errorf(fmt.Sprintf("Handler %s returned without error - got (%v) wanted (%v)", "StreamHandler", 200, 500))
		}
	})

	t.Run("Testing endpoint POST /api/v1/streamdata : should fail", func(t *testing.T) {
		req, _ = http.NewRequest("POST", "/api/v1/streamdata", bytes.NewBuffer([]byte("{\"error\"}")))
		rr := httptest.NewRecorder()
		conn := NewTestClient("../../tests/response.json", 500)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			StreamHandler(w, r, conn)
		})
		handler.ServeHTTP(rr, req)
		body, e := ioutil.ReadAll(rr.Body)
		if e != nil {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "StreamHandler", e, nil))
		}
		// ignore errors here
		json.Unmarshal(body, &response)
		conn.Debug(fmt.Sprintf("Response : %v", response))
		if rr.Code != 500 {
			t.Errorf(fmt.Sprintf("Handler %s returned without error - got (%v) wanted (%v)", "StreamHandler", 200, 500))
		}
	})

	t.Run("Testing endpoint POST /api/v1/streamdata : should pass", func(t *testing.T) {
		file, _ := ioutil.ReadFile("../../tests/segmentio.json")
		req, _ = http.NewRequest("POST", "/api/v1/streamdata", bytes.NewBuffer(file))
		rr := httptest.NewRecorder()
		conn := NewTestClient("../../tests/response.json", 200)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			StreamHandler(w, r, conn)
		})
		handler.ServeHTTP(rr, req)
		body, e := ioutil.ReadAll(rr.Body)
		if e != nil {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "StreamHandler", e, nil))
		}
		// ignore errors here
		err := json.Unmarshal(body, &response)
		conn.Debug(fmt.Sprintf("Response : %v", response))
		if rr.Code != 200 {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "StreamHandler", err, nil))
		}
	})

}
