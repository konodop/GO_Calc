package application

import (
	"os"
	"testing"
	"time"
)

func TestConfigFromEnv(t *testing.T) {
	os.Setenv("PORT", "1488")
	config := ConfigFromEnv()
	if config.Addr != "1488" {
		t.Errorf("Expected PORT to be 1488, got %s", config.Addr)
	}
}

func TestRunAgent(t *testing.T) {
	agent := New()

	go func() {
		if err := agent.RunServer(); err != nil {
			t.Errorf("RunServer() failed with error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
}
