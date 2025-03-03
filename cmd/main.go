package main

import (
	"calc/internal/application"
	"calc/pkg/agent"
)

func main() {
	app := application.New()
	Agent := agent.New()
	go Agent.RunAgent()
	app.RunServer()
}
