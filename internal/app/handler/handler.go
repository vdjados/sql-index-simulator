package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// GetServices обрабатывает главную страницу "/" и реализует фильтрацию по параметру filter.
// Фильтрация выполняется по полям Name и TableSize модели Service.
func (h *Handler) GetServices(ctx *gin.Context) {
	filter := ctx.Query("filter")

	services, err := h.Repository.GetServices(filter)
	if err != nil {
		logrus.Error(err)
	}

	currentRequest, err := h.Repository.GetCurrentRequest()
	if err != nil {
		logrus.Error(err)
	}

	cartCount := 0
	if currentRequest != nil {
		cartCount = len(currentRequest.Services)
	}

	ctx.HTML(http.StatusOK, "services.html", gin.H{
		"services":     services,
		"filter":       filter,
		"currentQuery": filter,
		"request":      currentRequest,
		"cartCount":    cartCount,
	})
}

// GetService показывает детальную информацию об одном индексе по его ID.
func (h *Handler) GetService(ctx *gin.Context) {
	id := ctx.Param("id")

	service, err := h.Repository.GetService(id)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, fmt.Sprintf("service %s not found", id))
		return
	}

	ctx.HTML(http.StatusOK, "service.html", gin.H{
		"service": service,
	})
}

// GetRequest показывает состав запроса: селективность, готовый результат и список индексов.
func (h *Handler) GetRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	req, err := h.Repository.GetRequest(id)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, fmt.Sprintf("request %s not found", id))
		return
	}

	ctx.HTML(http.StatusOK, "request.html", gin.H{
		"request": req,
	})
}
