package validator

import (
	"fmt"
	"os"
	"testing"

	"github.com/microlib/simple"
)

var (
	logger simple.Logger
)

func TestEnvars(t *testing.T) {
	logger := &simple.Logger{Level: "info"}

	t.Run("ValidateEnvars : should fail", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "")
		err := ValidateEnvars(logger)
		if err == nil {
			t.Errorf(fmt.Sprintf("Handler %s returned with no error - got (%v) wanted (%v)", "ValidateEnvars", err, nil))
		}
	})

	t.Run("ValidateEnvars : should pass", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "info")
		os.Setenv("ADV_URL", "/test")
		os.Setenv("SERVER_PORT", "9000")
		os.Setenv("REDIS_HOST", "127.0.0.1")
		os.Setenv("REDIS_PORT", "6379")
		os.Setenv("REDIS_PASSWORD", "6379")
		os.Setenv("MONGODB_HOST", "localhost")
		os.Setenv("MONGODB_PORT", "27017")
		os.Setenv("MONGODB_DATABASE", "test")
		os.Setenv("MONGODB_USER", "mp")
		os.Setenv("MONGODB_PASSWORD", "mp")
		os.Setenv("URL", "http://test.com")
		os.Setenv("TOKEN", "dsafsdfdsf")
		os.Setenv("VERSION", "1.0.3")
		os.Setenv("KAFKA_BROKERS", "localhost:9092")
		os.Setenv("TOPIC", "test")
		os.Setenv("CONNECTOR", "NA")
		os.Setenv("PROVIDER_NAME", "NA")
		os.Setenv("PROVIDER_URL", "http://test.com")
		os.Setenv("PROVIDER_TOKEN", "dsfgsdfsdf")
		os.Setenv("ANALYTICS_URL", "http://test.com")
		err := ValidateEnvars(logger)
		if err != nil {
			t.Errorf(fmt.Sprintf("Handler %s returned with error - got (%v) wanted (%v)", "ValidateEnvars", err, nil))
		}
	})
}
