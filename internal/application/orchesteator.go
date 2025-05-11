package application

import (
	"calc/pkg/agent"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/mattn/go-sqlite3"
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
	db, err      = sql.Open("sqlite3", "store.db")
	mySigningKey = []byte("secret")
	ctx          = context.TODO()
	mu           sync.Mutex
	tasks        = make(map[int]agent.Task)
	reses        = make(map[int]agent.Data)
	Addtime      = 500 * time.Millisecond
	Subtime      = 500 * time.Millisecond
	Multime      = 1 * time.Second
	Divtime      = 1 * time.Second
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

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type FullExpression struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	Result     string `json:"result"`
	Expression string `json:"expression"`
	UserID     int64  `json:"userid"`
}

type Request struct {
	Expression string `json:"expression"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Создание токена
func CreateToken(login string) (string, error) {
	claims := jwt.MapClaims{}
	claims["login"] = login
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(mySigningKey)

	return tokenString, err
}

// Middleware для проверки токена
func TokenVerifyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверка метода подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return mySigningKey, nil
		})

		person, err := getUserid(ctx, db, tokenString)

		if err != nil || !token.Valid || person == 0 {
			fmt.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	login := req.Login
	password := req.Password

	if len(login) >= 4 && len(password) >= 4 {
		var q = `
		SELECT id FROM users WHERE login = $1
		`
		var ok int64
		db.QueryRowContext(ctx, q, login).Scan(&ok)
		if ok != 0 {
			http.Error(w, "This login is already used", http.StatusConflict)
			return
		}

		token, err := CreateToken(login)

		q = `
		INSERT INTO users (login, password, token) values ($1, $2, $3)
		`
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Could not create token", http.StatusInternalServerError)
			return
		}
		_, err = db.ExecContext(ctx, q, login, password, token)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Could not create token", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(token))
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	login := req.Login
	password := req.Password

	var id int64
	var q = `
		SELECT id FROM users WHERE login = $1 AND password = $2
	`
	db.QueryRowContext(ctx, q, login, password).Scan(&id)
	if id != 0 { // Если пользователь существует
		token, err := CreateToken(login)
		if err != nil {
			http.Error(w, "Could not create token", http.StatusInternalServerError)
			return
		}
		var q = `
			UPDATE users SET token = $1 WHERE login = $2
		`
		_, err = db.ExecContext(ctx, q, token, login)
		if err != nil {
			http.Error(w, "Could not update token", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(token))
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

// Проверка работает ли система авторизации
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Authorized"))
}

func createTables(ctx context.Context, db *sql.DB) error {
	const (
		usersTable = `
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			login TEXT NOT NULL,
			password TEXT NOT NULL,
			token TEXT NOT NULL
		);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		status TEXT,
		result TEXT,
		expression TEXT NOT NULL,
		user_id INTEGER NOT NULL
	);`
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		return err
	}

	return nil
}

func insertExpression(ctx context.Context, db *sql.DB, expression *FullExpression) (int64, error) {
	var q = `
	INSERT INTO expressions (status, result, expression, user_id) values ($1, $2, $3, $4)
	`
	result, err := db.ExecContext(ctx, q, expression.Status, expression.Result, expression.Expression, expression.UserID)
	if err != nil {
		return 0, err
	}
	iD, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return iD, nil
}

func update_result(ctx context.Context, db *sql.DB, expression *FullExpression) error {
	var q = `
	UPDATE expressions SET status = $1, result = $2 WHERE id = $3
	`
	_, err := db.ExecContext(ctx, q, expression.Status, expression.Result, expression.ID)
	if err != nil {
		return err
	}

	return nil
}

func getUserid(ctx context.Context, db *sql.DB, token string) (int64, error) {
	var q = `
	SELECT id FROM users WHERE token = $1
	`
	var user_id int64
	err := db.QueryRowContext(ctx, q, token).Scan(&user_id)
	if err != nil {
		return 0, err
	}

	return user_id, nil
}

func getExpression(ctx context.Context, db *sql.DB, expression_id int64) (*FullExpression, error) {
	var q = `
	SELECT * FROM expressions WHERE id = $1
	`
	var expression FullExpression
	err := db.QueryRowContext(ctx, q, expression_id).Scan(&expression.ID, &expression.Status, &expression.Result, &expression.Expression, &expression.UserID)
	if err != nil {
		return nil, err
	}

	return &expression, nil
}

func getExpressions(ctx context.Context, db *sql.DB, user_id int64) ([]FullExpression, error) {
	var q = `
	SELECT * FROM expressions WHERE user_id = $1
	`
	var expressions []FullExpression
	rows, err := db.QueryContext(ctx, q, user_id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		e := FullExpression{}
		err := rows.Scan(&e.ID, &e.Status, &e.Result, &e.Expression, &e.UserID)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	return expressions, nil
}

func Calc(express string, idt int) {
	status := "started"
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
			status = "Cannot divide by 0"
			express = "-1"
			break
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
	if status == "started" {
		status = "ended"
	}
	mu.Lock()
	expres, err := getExpression(ctx, db, int64(idt))
	mu.Unlock()
	if err != nil {
		panic(err)
	}
	res := strconv.FormatFloat(result, 'f', 6, 64)

	ex := &FullExpression{
		ID:     expres.ID,
		Status: status,
		Result: res,
	}
	mu.Lock()
	err = update_result(ctx, db, ex)
	mu.Unlock()
	if err != nil {
		panic(err)
	}
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
	token := r.Header.Get("Authorization")

	usid, err := getUserid(ctx, db, token)

	if err != nil {
		panic(err)
	}

	ex := &FullExpression{
		Status:     "started",
		Result:     "NULL",
		Expression: express,
		UserID:     usid,
	}
	mu.Lock()
	expressionID, err := insertExpression(ctx, db, ex)
	mu.Unlock()
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	response := Responsetrue{
		Id: fmt.Sprintf("%d", expressionID),
	}
	json.NewEncoder(w).Encode(response)
	go Calc(express, int(expressionID))
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

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM expressions").Scan(&count)
		if err != nil {
			panic(err)
		}

		if idr > count || idr <= 0 {
			response := BadResponse{
				Result: "Not Found",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(response)
			return
		}
		mu.Lock()
		expression, _ := getExpression(ctx, db, int64(idr))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(expression)
	} else {
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()

		token := r.Header.Get("Authorization")

		usid, err := getUserid(ctx, db, token)

		if err != nil {
			panic(err)
		}

		expressions, _ := getExpressions(ctx, db, usid)
		mu.Unlock()
		var exprs []Expression
		for _, expression := range expressions {
			a := Expression{
				ID:     int(expression.ID),
				Status: expression.Status,
				Result: expression.Result,
			}
			exprs = append(exprs, a)

		}
		response := map[string][]Expression{"expressions": exprs}
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
	reses[result.ID] = result
	mu.Unlock()
}

func (a *Application) RunServer() error {
	fmt.Println("Оркестратор запускается")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.PingContext(ctx)
	if err != nil {
		panic(err)
	}

	if err = createTables(ctx, db); err != nil {
		panic(err)
	}

	// Защищенный маршрут
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/login", LoginHandler)
	mux.HandleFunc("/api/v1/register", RegisterHandler)
	mux.Handle("/ping", TokenVerifyMiddleware(http.HandlerFunc(PingHandler)))
	mux.Handle("/api/v1/calculate", TokenVerifyMiddleware(http.HandlerFunc(CalcHandler)))
	mux.Handle("/api/v1/expressions", TokenVerifyMiddleware(http.HandlerFunc(ExpressionsHandeler)))
	mux.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
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
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(response)
		}
	})

	fmt.Println("оркестратор запущен")
	return http.ListenAndServe(":8080", mux)
}
