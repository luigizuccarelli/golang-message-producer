package main

// Response schema
type Response struct {
	StatusCode string `json:"statuscode"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

type ConnectionData struct {
	Name    string
	Port    string
	Brokers string
	Topic   string
}
