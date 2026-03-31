package api

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/dsn"
	"web_backend/internal/app/handler"
	"web_backend/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server")

	_ = godotenv.Load() // загружает .env (DB_PASS, DB_HOST и т.д.)

	router := gin.Default()

	postgresString := dsn.FromEnv()
	logrus.Info("DSN: ", postgresString)

	repo, err := repository.NewRepository(postgresString)
	if err != nil {
		logrus.Error("Ошибка инициализации репозитория: ", err)
		return
	}

	handler := handler.NewHandler(repo)

	handler.RegisterAPI(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		logrus.Error("Ошибка запуска сервера: ", err)
	}
	log.Println("Server down")
}
