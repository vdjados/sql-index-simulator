package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"web_backend/docs"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Host = "localhost:" + port
	docs.SwaggerInfo.Title = "SQL Index Simulator API"
	docs.SwaggerInfo.Description = "Lab 4 API"

	handler.RegisterAPI(router)
	swaggerURL := ginSwagger.URL("/swagger/doc.json")
	router.Any("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))

	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	if err := router.Run(":" + port); err != nil {
		logrus.Error("Ошибка запуска сервера: ", err)
	}
	log.Println("Server down")
}
