package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/repository"
	"web_backend/internal/app/serializer"
)

type Handler struct {
	Repository *repository.Repository
}

type sqlQueriesListResponse struct {
	Total int                       `json:"total"`
	Items []serializer.SqlQueryJSON `json:"items"`
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")

	indexedTables := api.Group("/indexed-tables")
	{
		indexedTables.GET("", h.ApiGetServices)
		indexedTables.GET("/:id", h.ApiGetService)
		indexedTables.POST("", h.ApiCreateService)
	}

	sqlQueries := api.Group("/sql-queries")
	{
		sqlQueries.GET("/cart", h.ApiGetCart)
		sqlQueries.GET("", h.ApiGetSqlQueries)
		sqlQueries.GET("/:id", h.ApiGetSqlQuery)
		sqlQueries.PUT("/:id", h.ApiEditSqlQuery)
		sqlQueries.PUT("/:id/form", h.ApiFormSqlQuery)
		sqlQueries.PUT("/:id/finish", h.ApiFinishSqlQuery)
		sqlQueries.DELETE("/:id", h.ApiDeleteSqlQuery)
		sqlQueries.POST("/draft/indexed-tables/:indexed_table_id", h.ApiAddToDraft)
		sqlQueries.PUT("/:id/indexed-tables/:indexed_table_id", h.ApiEditItem)
		sqlQueries.DELETE("/:id/indexed-tables/:indexed_table_id", h.ApiDeleteItem)
	}

	users := api.Group("/users")
	{
		users.POST("/register", h.ApiRegisterUser)
		users.POST("/login", h.ApiLoginStub)
		users.POST("/logout", h.ApiLogoutStub)
	}
}

func (h *Handler) apiError(ctx *gin.Context, code int, err error) {
	logrus.Error(err)
	msg := err.Error()
	switch {
	case errors.Is(err, repository.ErrNotFound):
		msg = "Не найдено"
	case errors.Is(err, repository.ErrNotAllowed):
		msg = "Недостаточно прав"
	case errors.Is(err, repository.ErrNoDraft):
		msg = "Черновик не найден"
	case errors.Is(err, repository.ErrValidation):
		msg = "Некорректные данные"
	}
	ctx.JSON(code, gin.H{"status": "error", "description": msg})
}

// ===== API: Services =====

func (h *Handler) ApiGetServices(ctx *gin.Context) {
	filter := ctx.Query("filter")
	services, err := h.Repository.GetServices(filter)
	if err != nil {
		h.apiError(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := make([]serializer.ServiceJSON, 0, len(services))
	for _, s := range services {
		resp = append(resp, serializer.ServiceToJSON(s))
	}
	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) ApiGetService(ctx *gin.Context) {
	id := ctx.Param("id")
	s, err := h.Repository.GetService(id)
	if err != nil {
		h.apiError(ctx, http.StatusNotFound, repository.ErrNotFound)
		return
	}
	ctx.JSON(http.StatusOK, serializer.ServiceToJSON(s))
}

func (h *Handler) ApiCreateService(ctx *gin.Context) {
	contentType := ctx.GetHeader("Content-Type")
	var s repository.Service
	if strings.HasPrefix(contentType, "application/json") {
		if err := ctx.BindJSON(&s); err != nil {
			h.apiError(ctx, http.StatusBadRequest, err)
			return
		}
	} else {
		s.ID = ctx.PostForm("id")
		s.Name = ctx.PostForm("name")
		s.TableSize = ctx.PostForm("table_size")
		s.Speed = ctx.PostForm("speed")
		s.Description = ctx.PostForm("description")
	}
	if s.ID == "" || s.Name == "" || s.TableSize == "" {
		h.apiError(ctx, http.StatusBadRequest, repository.ErrValidation)
		return
	}
	s.Status = "active"
	if err := h.Repository.CreateService(s); err != nil {
		h.apiError(ctx, http.StatusInternalServerError, err)
		return
	}
	img, _ := ctx.FormFile("image")
	vid, _ := ctx.FormFile("video")
	if img != nil || vid != nil {
		updated, err := h.Repository.AddServiceMedia(ctx, s.ID, img, vid)
		if err != nil {
			h.apiError(ctx, http.StatusInternalServerError, err)
			return
		}
		s = updated
	}
	ctx.Header("Location", fmt.Sprintf("/api/services/%s", s.ID))
	ctx.JSON(http.StatusCreated, serializer.ServiceToJSON(s))
}

// ===== API: Cart & SqlQueries (scaffold) =====

func (h *Handler) ApiGetCart(ctx *gin.Context) {
	userID := repository.CreatorUserID()
	req, _ := h.Repository.GetCurrentRequest(userID)
	if req == nil {
		ctx.JSON(http.StatusOK, serializer.CartJSON{ID: nil, Count: 0})
		return
	}
	ctx.JSON(http.StatusOK, serializer.CartJSON{ID: &req.ID, Count: len(req.Services)})
}

func (h *Handler) ApiGetSqlQueries(ctx *gin.Context) {
	// фильтрация: status, from-date, to-date (как в примере)
	status := ctx.Query("status")
	fromStr := ctx.Query("from-date")
	toStr := ctx.Query("to-date")
	if fromStr == "" {
		fromStr = ctx.Query("formed-from")
	}
	if toStr == "" {
		toStr = ctx.Query("formed-to")
	}
	var from, to time.Time
	var err error
	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			h.apiError(ctx, http.StatusBadRequest, err)
			return
		}
	}
	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			h.apiError(ctx, http.StatusBadRequest, err)
			return
		}
	}
	list, err := h.Repository.ApiListSqlQueries(from, to, status)
	if err != nil {
		h.apiError(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := make([]serializer.SqlQueryJSON, 0, len(list))
	for _, q := range list {
		creator, moderator, _ := h.Repository.GetModeratorAndCreatorLogin(q)
		completedCount := h.Repository.GetCompletedItemCount(q.ID)
		resp = append(resp, serializer.SqlQueryToJSON(q, creator, moderator, completedCount))
	}
	ctx.JSON(http.StatusOK, sqlQueriesListResponse{
		Total: len(resp),
		Items: resp,
	})
}

func (h *Handler) ApiGetSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	q, err := h.Repository.ApiGetSqlQuery(uint(idUint64))
	if err != nil {
		h.apiError(ctx, http.StatusNotFound, err)
		return
	}
	creator, moderator, _ := h.Repository.GetModeratorAndCreatorLogin(q)
	completedCount := h.Repository.GetCompletedItemCount(q.ID)
	items := make([]serializer.SqlQueryItemJSON, 0, len(q.Services))
	for _, it := range q.Services {
		items = append(items, serializer.SqlQueryItemToJSON(it))
	}
	ctx.JSON(http.StatusOK, gin.H{
		"sql_query": serializer.SqlQueryToJSON(q, creator, moderator, completedCount),
		"items":     items,
	})
}

type editSqlQueryBody struct {
	QueryDescription *string  `json:"query_description"`
	QueryText        *string  `json:"query_text"`
	Theme            *string  `json:"theme"`
	Selectivity      *float64 `json:"selectivity"`
}

func (h *Handler) ApiEditSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	var body editSqlQueryBody
	if err := ctx.BindJSON(&body); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	queryDescription := body.QueryDescription
	if queryDescription == nil {
		queryDescription = body.QueryText
	}
	if queryDescription == nil {
		queryDescription = body.Theme
	}
	q, err := h.Repository.ApiEditSqlQuery(repository.CreatorUserID(), uint(idUint64), queryDescription, body.Selectivity)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	creator, moderator, _ := h.Repository.GetModeratorAndCreatorLogin(q)
	completedCount := h.Repository.GetCompletedItemCount(q.ID)
	ctx.JSON(http.StatusOK, serializer.SqlQueryToJSON(q, creator, moderator, completedCount))
}

func (h *Handler) ApiFormSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	q, err := h.Repository.ApiFormSqlQuery(repository.CreatorUserID(), uint(idUint64))
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	creator, moderator, _ := h.Repository.GetModeratorAndCreatorLogin(q)
	completedCount := h.Repository.GetCompletedItemCount(q.ID)
	ctx.JSON(http.StatusOK, serializer.SqlQueryToJSON(q, creator, moderator, completedCount))
}

type statusBody struct {
	Status string `json:"status"`
}

func (h *Handler) ApiFinishSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	var body statusBody
	if err := ctx.BindJSON(&body); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	q, err := h.Repository.ApiFinishSqlQuery(uint(idUint64), body.Status)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	creator, moderator, _ := h.Repository.GetModeratorAndCreatorLogin(q)
	completedCount := h.Repository.GetCompletedItemCount(q.ID)
	ctx.JSON(http.StatusOK, serializer.SqlQueryToJSON(q, creator, moderator, completedCount))
}

func (h *Handler) ApiDeleteSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.ApiDeleteSqlQuery(repository.CreatorUserID(), uint(idUint64)); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "sql_query удалён"})
}

// ===== API: m-m =====

func (h *Handler) ApiAddToDraft(ctx *gin.Context) {
	serviceID := ctx.Param("indexed_table_id")
	item, err := h.Repository.ApiAddToDraft(repository.CreatorUserID(), serviceID)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusCreated, serializer.SqlQueryItemToJSON(item))
}

type editItemBody struct {
	Quantity    *int     `json:"quantity"`
	Position    *int     `json:"position"`
	Selectivity *float64 `json:"selectivity"`
}

func (h *Handler) ApiEditItem(ctx *gin.Context) {
	serviceID := ctx.Param("indexed_table_id")
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	var body editItemBody
	if err := ctx.BindJSON(&body); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	item, err := h.Repository.ApiEditItem(repository.CreatorUserID(), uint(idUint64), serviceID, body.Quantity, body.Position, body.Selectivity)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.SqlQueryItemToJSON(item))
}

func (h *Handler) ApiDeleteItem(ctx *gin.Context) {
	serviceID := ctx.Param("indexed_table_id")
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.ApiDeleteItem(repository.CreatorUserID(), uint(idUint64), serviceID); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "позиция удалена"})
}

// ===== API: users =====

func (h *Handler) ApiRegisterUser(ctx *gin.Context) {
	var body serializer.RegisterUserJSON
	if err := ctx.BindJSON(&body); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	u, err := h.Repository.CreateUser(repository.User{Name: strings.TrimSpace(body.Name), Email: strings.TrimSpace(body.Email), Role: "user"})
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.Header("Location", fmt.Sprintf("/api/users/%d", u.ID))
	ctx.JSON(http.StatusCreated, serializer.UserToJSON(u))
}
func (h *Handler) ApiLoginStub(ctx *gin.Context)  { ctx.JSON(http.StatusOK, gin.H{"status": "ok"}) }
func (h *Handler) ApiLogoutStub(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"status": "ok"}) }

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
// Если индекс не найден — редирект на главную.
func (h *Handler) GetService(ctx *gin.Context) {
	id := ctx.Param("id")

	service, err := h.Repository.GetService(id)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, "/")
		return
	}

	ctx.HTML(http.StatusOK, "service.html", gin.H{
		"service": service,
	})
}

// GetSqlQuery показывает состав sql_query: селективность, результат (время и память) и список индексов.
func (h *Handler) GetSqlQuery(ctx *gin.Context) {
	idParam := ctx.Param("id")
	idUint64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/")
		return
	}
	id := uint(idUint64)

	req, err := h.Repository.GetRequest(id)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, "/")
		return
	}

	ctx.HTML(http.StatusOK, "sql_query.html", gin.H{
		"sql_query": req,
	})
}

// AddToSqlQuery добавляет индекс (услугу) в текущий sql_query (черновик) пользователя через ORM.
func (h *Handler) AddToSqlQuery(ctx *gin.Context) {
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

// DeleteSqlQuery удаляет sql_query через raw SQL UPDATE (без ORM).
func (h *Handler) DeleteSqlQuery(ctx *gin.Context) {
	idParam := ctx.Param("id")
	idUint64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/")
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
