package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

type Expression struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Result string `json:"result"`
}

type Response struct {
	Expressions []Expression `json:"expressions"`
}

func FrontHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://localhost:8080/api/v1/expressions")
	if err != nil {
		http.Error(w, "Не получилось получить данные", 500)
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var response Response

	json.Unmarshal(body, &response)

	tmpl := template.Must(template.ParseFiles("html.html"))
	tmpl.Execute(w, response.Expressions)
}

func main() {
	http.HandleFunc("/", FrontHandler)

	fmt.Println("Фронт запущен на :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Println(err)
	}
}
