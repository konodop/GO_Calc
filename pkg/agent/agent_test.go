package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestConfigFromEnv(t *testing.T) {
	os.Setenv("COMPUTING_POWER", "4")
	config := ConfigFromEnv()
	if config.COMPUTING_POWER != 4 {
		t.Errorf("Expected COMPUTING_POWER to be 4, got %d", config.COMPUTING_POWER)
	}

	os.Setenv("COMPUTING_POWER", "-1")
	config = ConfigFromEnv()
	if config.COMPUTING_POWER != 8 {
		t.Errorf("Expected COMPUTING_POWER to be 8 for invalid input, got %d", config.COMPUTING_POWER)
	}
}

func TestResponder(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		task := Task{ID: 1, Arg1: 5, Arg2: 3, Operation: '+'}
		json.NewEncoder(w).Encode(task)
	}))
	defer ts.Close()

	go Responder()

	time.Sleep(100 * time.Millisecond)

}
func TestRunAgent(t *testing.T) {
	agent := New()

	go func() {
		if err := agent.RunAgent(); err != nil {
			t.Errorf("RunAgent() failed with error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
}
