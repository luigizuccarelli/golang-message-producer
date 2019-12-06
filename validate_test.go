package main

import (
	"fmt"
	"github.com/microlib/simple"
	"os"
	"testing"
)

var (
	logger simple.Logger
)

func TestEnvars(t *testing.T) {

	// create anonymous struct
	tests := []struct {
		Name     string
		Payload  string
		Handler  string
		FileName string
		Want     bool
		ErrorMsg string
	}{
		{
			"Test envars : should fail",
			"",
			"TestEnvarsFail",
			"",
			true,
			"Handler %s returned - got (%v) wanted (%v)",
		},
		{
			"Test envars : should pass",
			"",
			"TestEnvarsPass",
			"",
			false,
			"Handler %s returned - got (%v) wanted (%v)",
		},
	}
	var err error
	logger.Level = "trace"
	for _, tt := range tests {
		fmt.Println(fmt.Sprintf("\nExecuting test : %s", tt.Name))
		switch tt.Handler {
		case "TestEnvarsFail":
			err = nil
			os.Setenv("SERVER_PORT", "")
			err = ValidateEnvars()
		case "TestEnvarsPass":
			err = nil
			os.Setenv("SERVER_PORT", "9000")
			os.Setenv("REDIS_HOST", "127.0.0.1")
			os.Setenv("REDIS_PORT", "6379")
			os.Setenv("REDIS_PASSWORD", "6379")
			os.Setenv("URL", "http://test.com")
			os.Setenv("TOKEN", "dsafsdfdsf")
			os.Setenv("VERSION", "1.0.3")
			os.Setenv("BROKERS", "localhost:9092")
			os.Setenv("TOPIC", "test")
			os.Setenv("CONNECTOR", "NA")
			err = ValidateEnvars()
		}

		if !tt.Want {
			if err != nil {
				t.Errorf(fmt.Sprintf(tt.ErrorMsg, tt.Handler, err, nil))
			}
		} else {
			if err == nil {
				t.Errorf(fmt.Sprintf(tt.ErrorMsg, tt.Handler, "nil", "error"))
			}
		}
		fmt.Println("")
	}
}
