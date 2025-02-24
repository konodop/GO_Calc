package application

import (
	"calc/pkg/agent"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Addr string
}

type BadResponse struct {
	Result string `json:"error"`
}

type Response struct {
	Result string `json:"result"`
}

type Responsetrue struct {
	Id string `json:"id"`
}

// глобальные переменные

var (
	id          int = 0
	expressions []Expression
	mu          sync.Mutex
	// tasks       []agent.Task
)

func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

type Application struct {
	config *Config
}

func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

type Expression struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Result string `json:"result"`
}

type Request struct {
	Expression string `json:"expression"`
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	request := new(Request)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response := BadResponse{
			Result: "Expression is not valid",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(response)
		return
	}
	request.Expression = strings.ReplaceAll(request.Expression, " ", "")
	re := regexp.MustCompile(`[^0-9\-+/*()]`)
	if re.MatchString(request.Expression) {
		response := BadResponse{
			Result: "Expression is not valid",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(response)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	id++
	response := Responsetrue{
		Id: fmt.Sprintf("%d", id),
	}
	json.NewEncoder(w).Encode(response)
	// aaaaaaaaaaaaaaaaaaaa
	express := request.Expression
	opn := 0
	cls := 0
	if len(express) == 0 {
		response := BadResponse{
			Result: "Expression is not valid",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(response)
		return
	}
	mu.Lock()
	expressions = append(expressions, Expression{ID: id, Status: "started", Result: "NULL"})
	mu.Unlock()
	bad_res := 0
	for i, j := range express {
		//проверка на скобки
		if j == '(' {
			if i+1 < len(express) {
				if express[i+1] == ')' {
					bad_res = 1
					break
				}
			}
			opn++
		} else if j == ')' {
			cls++
		}
		if opn < cls {
			bad_res = 1
			break
		}
		// ищем высший знак
		if j == '+' || j == '*' || j == '/' {
			if i == 0 || i == len(express)-1 {
				bad_res = 1
				break
			} else if express[i-1] == '(' || express[i+1] == '+' || express[i+1] == '*' || express[i+1] == '/' || express[i+1] == ')' {
				bad_res = 1
				break
			}

		} else if j == '-' {
			if i == len(express)-1 {
				bad_res = 1
				break
			} else if express[i+1] == '-' || express[i+1] == '+' || express[i+1] == '*' || express[i+1] == '/' || express[i+1] == ')' {
				bad_res = 1
				break
			}
		}
	}

	if opn != cls {
		bad_res = 1
	}

	if bad_res == 1 {
		response := BadResponse{
			Result: "Expression is not valid",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(response)
		return
	}

	for {
		var s int = 0
		for i := range len(express) {
			if express[i] == '(' {
				s = i + 1
			}
		}
		var q rune = ' '
		i := s
		index := 0
		for i < len(express) && express[i] != ')' {
			if express[i] == '*' {
				index = i
				q = '*'
				break
			} else if express[i] == '/' {
				index = i
				q = '/'
				break
			} else if express[i] == '+' && q == ' ' {
				q = '+'
				index = i
			} else if express[i] == '-' && q == ' ' {
				q = '-'
				index = i
			}
			i++
		}
		if q == ' ' {
			if s != 0 {
				if express[s-1] == '(' && i < len(express) {
					express = express[:s-1] + express[s:i] + express[i+1:]
					continue
				}
			}
			break
		}
		right := ""
		left := ""
		i = index - 1
		for i >= 0 {
			g := express[i]
			if g == '(' {
				break
			} else if g == '+' || g == '*' || g == '/' {
				break
			} else if g == '-' {
				if i == s {
					left = "-" + left
					i--
				} else if express[i-1] == '-' || express[i-1] == '+' || express[i-1] == '*' || express[i-1] == '/' {
					left = "-" + left
					i--
				}
				break
			}
			left = express[i:i+1] + left
			i--
		}
		lt := i + 1
		i = index + 1
		for i < len(express) {
			g := express[i]
			if g == ')' {
				break
			} else if g == '+' || g == '*' || g == '/' {
				break
			} else if g == '-' {
				if express[i-1] == '-' || express[i-1] == '+' || express[i-1] == '*' || express[i-1] == '/' {
				} else {
					break
				}
			}
			right = right + express[i:i+1]
			i++
		}
		rt := i
		lefted, _ := strconv.ParseFloat(left, 64)
		righted, _ := strconv.ParseFloat(right, 64)
		if q == '/' && righted == 0 {
			response := BadResponse{
				Result: "Cannot divide by zero",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(422)
			json.NewEncoder(w).Encode(response)
			return
		} else if q == '-' && len(left) == 0 {
			break
		}

		f := agent.Task{ID: id - 1, Arg1: lefted, Arg2: righted, Operation: q, Operation_time: time.Duration(1 * time.Second)}
		result := agent.Calc(f)
		express = express[:lt] + result.Result + express[rt:]
	}

	result, _ := strconv.ParseFloat(express, 64)
	mu.Lock()
	expressions[id-1].Status = "ended"
	expressions[id-1].Result = strconv.FormatFloat(result, 'f', 6, 64)
	mu.Unlock()
}

func ExpressionsHandeler(w http.ResponseWriter, r *http.Request) {
	idr := r.URL.Query().Get("id")
	if idr != "" {
		idr, err := strconv.Atoi(idr)
		if err != nil {
			response := BadResponse{
				Result: "Expression is not valid",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(422)
			json.NewEncoder(w).Encode(response)
			return
		}
		if idr > id || idr <= 0 {
			response := BadResponse{
				Result: "Not Found",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(response)
			return
		}
		mu.Lock()
		expression := expressions[idr-1]
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(expression)
	} else {
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		response := map[string][]Expression{"expressions": expressions}
		mu.Unlock()
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(response)
	}
}

func TaskHandeler(w http.ResponseWriter, r *http.Request) {
	response := BadResponse{
		Result: "Internal Server Error",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	json.NewEncoder(w).Encode(response)
}

func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	http.HandleFunc("/api/v1/expressions", ExpressionsHandeler)
	http.HandleFunc("/intrenal/task", TaskHandeler)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}
