package main

import (
	"log"
	"web_backend/internal/api"

	_ "web_backend/docs"
)

// @title SQL Index Simulator API
// @version 1.0
// @description API для лабораторной 4: JWT авторизация, роли, Redis blacklist
// @host localhost:8082
// @BasePath /api
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}
