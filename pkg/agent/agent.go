package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Task struct {
	ID             int           `json:"id"`
	Arg1           float64       `json:"arg1"`
	Arg2           float64       `json:"arg2"`
	Operation      rune          `json:"operation"`
	Operation_time time.Duration `json:"operation_time"`
}

type Data struct {
	ID     int    `json:"id"`
	Result string `json:"result"`
}

type Config struct {
	COMPUTING_POWER int
}

func ConfigFromEnv() *Config {
	config := new(Config)
	config.COMPUTING_POWER, _ = strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if config.COMPUTING_POWER <= 0 {
		config.COMPUTING_POWER = 8
	}
	return config
}

var comp_power int
var threads int

type Agent struct {
	config *Config
}

func New() *Agent {
	return &Agent{
		config: ConfigFromEnv(),
	}
}

func Calc(task Task) {

	ans := ""
	lefted := task.Arg1
	righted := task.Arg2
	var q rune = ' '
	if q == '*' {
		ans = strconv.FormatFloat(lefted*righted, 'f', 6, 64)
	} else if q == '/' {
		ans = strconv.FormatFloat(lefted/righted, 'f', 6, 64)
	} else if q == '-' {
		ans = strconv.FormatFloat(lefted-righted, 'f', 6, 64)
	} else {
		ans = strconv.FormatFloat(lefted+righted, 'f', 6, 64)
	}
	f := Data{ID: task.ID, Result: ans}
	threads--
	body, _ := json.Marshal(f)
	resp, _ := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(body))
	resp.Body.Close()
}

func Responder() {
	var task Task
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil || resp.StatusCode == 404 {
		time.Sleep(100 * time.Millisecond)
		return
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&task)
	for {
		if threads <= comp_power {
			go Calc(task)
			threads++
			break
		}
	}
}

func (a *Agent) RunAgent() error {
	comp_power = a.config.COMPUTING_POWER
	for {
		Responder()
	}
}
