package api

import (
	"fmt"
	"html/template"
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

	router.SetFuncMap(template.FuncMap{
		"printf": fmt.Sprintf,
	})

	postgresString := dsn.FromEnv()
	logrus.Info("DSN: ", postgresString)

	repo, err := repository.NewRepository(postgresString)
	if err != nil {
		logrus.Error("Ошибка инициализации репозитория: ", err)
		return
	}

	handler := handler.NewHandler(repo)

	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")

	router.GET("/", handler.GetServices)
	router.GET("/service/:id", handler.GetService)
	router.GET("/request/:id", handler.GetRequest)
	router.POST("/request/add", handler.AddToRequest)
	router.POST("/request/:id/delete", handler.DeleteRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		logrus.Error("Ошибка запуска сервера: ", err)
	}
	log.Println("Server down")
}
