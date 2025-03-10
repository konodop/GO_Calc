package main

import (
	"calc/internal/application"
)

func main() {
	app := application.New()
	app.RunServer()
}
