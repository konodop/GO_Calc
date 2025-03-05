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
	Addr    string
	Addtime time.Duration
	Subtime time.Duration
	Multime time.Duration
	Divtime time.Duration
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
	tasks       = make(map[int]agent.Task)
	reses       = make(map[int]agent.Data)
	Addtime     = 500 * time.Millisecond
	Subtime     = 500 * time.Millisecond
	Multime     = 1 * time.Second
	Divtime     = 1 * time.Second
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

func Calc(express string, idt int) {
	for {
		var s int = 0
		for i := range len(express) {
			if express[i] == '(' {
				s = i + 1
			}
		}
		var q rune = ' '
		var dur time.Duration
		i := s
		index := 0
		for i < len(express) && express[i] != ')' {
			if express[i] == '*' {
				index = i
				dur = Multime
				q = '*'
				break
			} else if express[i] == '/' {
				index = i
				q = '/'
				dur = Divtime
				break
			} else if express[i] == '+' && q == ' ' {
				q = '+'
				dur = Addtime
				index = i
			} else if express[i] == '-' && q == ' ' {
				q = '-'
				dur = Subtime
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
			mu.Lock()
			expressions[id-1].Status = "Cannot divide by 0"
			mu.Unlock()
			return
		} else if q == '-' && len(left) == 0 {
			express = "-" + right
			break
		}
		f := agent.Task{ID: idt, Arg1: lefted, Arg2: righted, Operation: q, Operation_time: dur}
		mu.Lock()
		tasks[idt] = f
		mu.Unlock()
		r := true
		var result string
		for r {
			mu.Lock()
			for _, i := range reses {
				if i.ID == idt {
					result = i.Result
					delete(reses, i.ID)
					r = false
					break
				}
			}
			mu.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
		express = express[:lt] + result + express[rt:]
	}
	result, _ := strconv.ParseFloat(express, 64)
	mu.Lock()
	expressions[idt-1].Status = "ended"
	expressions[idt-1].Result = strconv.FormatFloat(result, 'f', 6, 64)
	mu.Unlock()
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	id++
	response := Responsetrue{
		Id: fmt.Sprintf("%d", id),
	}
	json.NewEncoder(w).Encode(response)

	mu.Lock()
	expressions = append(expressions, Expression{ID: id, Status: "started", Result: "NULL"})
	mu.Unlock()
	go Calc(express, id)
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

func getTask(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mu.Lock()
	defer mu.Unlock()
	if len(tasks) == 0 {
		response := BadResponse{
			Result: "Not Found",
		}
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(response)
		return
	}

	for _, i := range tasks {
		delete(tasks, i.ID)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(i)
		return
	}

}

func postResult(w http.ResponseWriter, r *http.Request) {
	var result agent.Data
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		response := BadResponse{
			Result: "Expression is not valid",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode("a")
	mu.Lock()
	reses[id] = result
	mu.Unlock()
}

func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	http.HandleFunc("/api/v1/expressions", ExpressionsHandeler)
	http.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTask(w, r)
		case http.MethodPost:
			postResult(w, r)
		default:
			response := BadResponse{
				Result: "Method Not Allowed",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(405)
			json.NewEncoder(w).Encode(response)
		}
	})
	return http.ListenAndServe(":"+a.config.Addr, nil)
}
