package handler

import (
	"fmt"
	"net/http"
	"strconv"

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

	// Для лабораторной работы используем фиксированного пользователя с ID = 1.
	const userID = 1

	currentRequest, err := h.Repository.GetCurrentRequest(userID)
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
	idParam := ctx.Param("id")
	idUint64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid request id %s", idParam))
		return
	}
	id := uint(idUint64)

	req, err := h.Repository.GetRequest(id)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, fmt.Sprintf("request %d not found", id))
		return
	}

	ctx.HTML(http.StatusOK, "request.html", gin.H{
		"request": req,
	})
}

// AddToRequest добавляет услугу в текущую заявку (черновик) пользователя через ORM.
func (h *Handler) AddToRequest(ctx *gin.Context) {
	serviceID := ctx.PostForm("service_id")
	if serviceID == "" {
		ctx.String(http.StatusBadRequest, "service_id is required")
		return
	}

	const userID = 1

	_, err := h.Repository.AddServiceToDraft(userID, serviceID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}

// DeleteRequest логически удаляет заявку через raw SQL UPDATE (без ORM).
func (h *Handler) DeleteRequest(ctx *gin.Context) {
	idParam := ctx.Param("id")
	idUint64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid request id %s", idParam))
		return
	}
	requestID := uint(idUint64)

	const userID = 1

	if err := h.Repository.DeleteRequestLogical(userID, requestID); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
