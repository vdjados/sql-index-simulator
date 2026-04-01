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

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterAPI godoc
// @title SQL Index Simulator API
// @version 1.0
// @description API для лабораторной 4: JWT авторизация, роли, Redis blacklist, заявки и индексы
// @host localhost:8082
// @BasePath /api
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func (h *Handler) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")

	publicIndexedTables := api.Group("/indexed-tables")
	{
		publicIndexedTables.GET("", h.ApiGetServices)
		publicIndexedTables.GET("/:id", h.ApiGetService)
	}

	publicUsers := api.Group("/users")
	{
		publicUsers.POST("/register", h.ApiRegisterUser)
		publicUsers.POST("/login", h.ApiLogin)
	}

	authenticated := api.Group("/")
	authenticated.Use(h.AuthRequired())
	{
		authenticated.POST("/users/logout", h.ApiLogout)

		sqlQueries := authenticated.Group("/sql-queries")
		{
			sqlQueries.GET("/cart", h.ApiGetCart)
			sqlQueries.GET("", h.ApiGetSqlQueries)
			sqlQueries.GET("/:id", h.ApiGetSqlQuery)
			sqlQueries.PUT("/:id", h.ApiEditSqlQuery)
			sqlQueries.PUT("/:id/form", h.ApiFormSqlQuery)
			sqlQueries.DELETE("/:id", h.ApiDeleteSqlQuery)
			sqlQueries.POST("/draft/indexed-tables/:indexed_table_id", h.ApiAddToDraft)
			sqlQueries.PUT("/:id/indexed-tables/:indexed_table_id", h.ApiEditItem)
			sqlQueries.DELETE("/:id/indexed-tables/:indexed_table_id", h.ApiDeleteItem)
		}
	}

	moderator := api.Group("/")
	moderator.Use(h.AuthRequired(), h.ModeratorOnly())
	{
		moderator.POST("/indexed-tables", h.ApiCreateService)
		moderator.PUT("/sql-queries/:id/finish", h.ApiFinishSqlQuery)
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

// ApiGetServices godoc
// @Summary Получить список индексов
// @Description Публичный список indexed tables с фильтрацией по строке
// @Tags indexed-tables
// @Produce json
// @Param filter query string false "Фильтр по name/table_size"
// @Success 200 {array} serializer.ServiceJSON
// @Failure 500 {object} serializer.ErrorResponse
// @Router /indexed-tables [get]
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

// ApiGetService godoc
// @Summary Получить индекс по ID
// @Tags indexed-tables
// @Produce json
// @Param id path string true "ID индекса"
// @Success 200 {object} serializer.ServiceJSON
// @Failure 404 {object} serializer.ErrorResponse
// @Router /indexed-tables/{id} [get]
func (h *Handler) ApiGetService(ctx *gin.Context) {
	id := ctx.Param("id")
	s, err := h.Repository.GetService(id)
	if err != nil {
		h.apiError(ctx, http.StatusNotFound, repository.ErrNotFound)
		return
	}
	ctx.JSON(http.StatusOK, serializer.ServiceToJSON(s))
}

// ApiCreateService godoc
// @Summary Создать индекс (moderator)
// @Tags indexed-tables
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param id formData string true "ID индекса"
// @Param name formData string true "Название"
// @Param table_size formData string true "Размер таблицы"
// @Param speed formData string false "Скорость"
// @Param description formData string false "Описание"
// @Param image formData file false "Картинка"
// @Param video formData file false "Видео/GIF"
// @Success 201 {object} serializer.ServiceJSON
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Failure 403 {object} serializer.ErrorResponse
// @Router /indexed-tables [post]
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
	ctx.Header("Location", fmt.Sprintf("/api/indexed-tables/%s", s.ID))
	ctx.JSON(http.StatusCreated, serializer.ServiceToJSON(s))
}

// ===== API: Cart & SqlQueries (scaffold) =====

// ApiGetCart godoc
// @Summary Иконка корзины текущего пользователя
// @Tags sql-queries
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} serializer.CartJSON
// @Failure 401 {object} serializer.ErrorResponse
// @Router /sql-queries/cart [get]
func (h *Handler) ApiGetCart(ctx *gin.Context) {
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	req, _ := h.Repository.GetCurrentRequest(user.ID)
	if req == nil {
		ctx.JSON(http.StatusOK, serializer.CartJSON{ID: nil, Count: 0})
		return
	}
	ctx.JSON(http.StatusOK, serializer.CartJSON{ID: &req.ID, Count: len(req.Services)})
}

// ApiGetSqlQueries godoc
// @Summary Список заявок
// @Description Creator видит только свои заявки, moderator видит все
// @Tags sql-queries
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "Статус"
// @Param formed-from query string false "Дата от YYYY-MM-DD"
// @Param formed-to query string false "Дата до YYYY-MM-DD"
// @Param from-date query string false "Старый алиас formed-from"
// @Param to-date query string false "Старый алиас formed-to"
// @Success 200 {object} serializer.SqlQueriesListResponse
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Router /sql-queries [get]
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
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	list, err := h.Repository.ApiListSqlQueriesForUser(user.ID, user.IsModerator(), from, to, status)
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
	ctx.JSON(http.StatusOK, serializer.SqlQueriesListResponse{
		Total: len(resp),
		Items: resp,
	})
}

// ApiGetSqlQuery godoc
// @Summary Получить одну заявку
// @Tags sql-queries
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} serializer.SqlQueryDetailsResponse
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Failure 404 {object} serializer.ErrorResponse
// @Router /sql-queries/{id} [get]
func (h *Handler) ApiGetSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	q, err := h.Repository.ApiGetSqlQueryForUser(uint(idUint64), user.ID, user.IsModerator())
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
	ctx.JSON(http.StatusOK, serializer.SqlQueryDetailsResponse{
		SqlQuery: serializer.SqlQueryToJSON(q, creator, moderator, completedCount),
		Items:    items,
	})
}

// ApiEditSqlQuery godoc
// @Summary Обновить поля заявки
// @Tags sql-queries
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID заявки"
// @Param body body serializer.EditSqlQueryJSON true "Поля заявки"
// @Success 200 {object} serializer.SqlQueryJSON
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Failure 403 {object} serializer.ErrorResponse
// @Router /sql-queries/{id} [put]
func (h *Handler) ApiEditSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	var body serializer.EditSqlQueryJSON
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
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	q, err := h.Repository.ApiEditSqlQuery(user.ID, uint(idUint64), queryDescription, body.Selectivity)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	creator, moderator, _ := h.Repository.GetModeratorAndCreatorLogin(q)
	completedCount := h.Repository.GetCompletedItemCount(q.ID)
	ctx.JSON(http.StatusOK, serializer.SqlQueryToJSON(q, creator, moderator, completedCount))
}

// ApiFormSqlQuery godoc
// @Summary Сформировать заявку
// @Tags sql-queries
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} serializer.SqlQueryJSON
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Failure 403 {object} serializer.ErrorResponse
// @Router /sql-queries/{id}/form [put]
func (h *Handler) ApiFormSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	q, err := h.Repository.ApiFormSqlQuery(user.ID, uint(idUint64))
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	creator, moderator, _ := h.Repository.GetModeratorAndCreatorLogin(q)
	completedCount := h.Repository.GetCompletedItemCount(q.ID)
	ctx.JSON(http.StatusOK, serializer.SqlQueryToJSON(q, creator, moderator, completedCount))
}

// ApiFinishSqlQuery godoc
// @Summary Завершить/отклонить заявку (только moderator)
// @Tags sql-queries
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID заявки"
// @Param body body serializer.StatusJSON true "Новый статус: completed/rejected"
// @Success 200 {object} serializer.SqlQueryJSON
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Failure 403 {object} serializer.ErrorResponse
// @Router /sql-queries/{id}/finish [put]
func (h *Handler) ApiFinishSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	var body serializer.StatusJSON
	if err := ctx.BindJSON(&body); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	q, err := h.Repository.ApiFinishSqlQuery(uint(idUint64), body.Status, user.ID)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	creator, moderator, _ := h.Repository.GetModeratorAndCreatorLogin(q)
	completedCount := h.Repository.GetCompletedItemCount(q.ID)
	ctx.JSON(http.StatusOK, serializer.SqlQueryToJSON(q, creator, moderator, completedCount))
}

// ApiDeleteSqlQuery godoc
// @Summary Удалить черновик заявки
// @Tags sql-queries
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} serializer.MessageResponse
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Failure 403 {object} serializer.ErrorResponse
// @Router /sql-queries/{id} [delete]
func (h *Handler) ApiDeleteSqlQuery(ctx *gin.Context) {
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	if err := h.Repository.ApiDeleteSqlQuery(user.ID, uint(idUint64)); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "sql_query удалён"})
}

// ===== API: m-m =====

// ApiAddToDraft godoc
// @Summary Добавить индекс в черновик
// @Tags sql-query-items
// @Produce json
// @Security ApiKeyAuth
// @Param indexed_table_id path string true "ID индекса"
// @Success 201 {object} serializer.SqlQueryItemJSON
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Router /sql-queries/draft/indexed-tables/{indexed_table_id} [post]
func (h *Handler) ApiAddToDraft(ctx *gin.Context) {
	serviceID := ctx.Param("indexed_table_id")
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	item, err := h.Repository.ApiAddToDraft(user.ID, serviceID)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusCreated, serializer.SqlQueryItemToJSON(item))
}

// ApiEditItem godoc
// @Summary Изменить позицию в заявке
// @Tags sql-query-items
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID заявки"
// @Param indexed_table_id path string true "ID индекса"
// @Param body body serializer.EditSqlQueryItemJSON true "Поля позиции"
// @Success 200 {object} serializer.SqlQueryItemJSON
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Failure 403 {object} serializer.ErrorResponse
// @Router /sql-queries/{id}/indexed-tables/{indexed_table_id} [put]
func (h *Handler) ApiEditItem(ctx *gin.Context) {
	serviceID := ctx.Param("indexed_table_id")
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	var body serializer.EditSqlQueryItemJSON
	if err := ctx.BindJSON(&body); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	item, err := h.Repository.ApiEditItem(user.ID, uint(idUint64), serviceID, body.Quantity, body.Position, body.Selectivity)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.SqlQueryItemToJSON(item))
}

// ApiDeleteItem godoc
// @Summary Удалить позицию из заявки
// @Tags sql-query-items
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID заявки"
// @Param indexed_table_id path string true "ID индекса"
// @Success 200 {object} serializer.MessageResponse
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Failure 403 {object} serializer.ErrorResponse
// @Router /sql-queries/{id}/indexed-tables/{indexed_table_id} [delete]
func (h *Handler) ApiDeleteItem(ctx *gin.Context) {
	serviceID := ctx.Param("indexed_table_id")
	idUint64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	if err := h.Repository.ApiDeleteItem(user.ID, uint(idUint64), serviceID); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "позиция удалена"})
}

// ===== API: users =====

// ApiRegisterUser godoc
// @Summary Регистрация
// @Tags users
// @Accept json
// @Produce json
// @Param body body serializer.RegisterUserJSON true "Пользователь"
// @Success 201 {object} serializer.UserJSON
// @Failure 400 {object} serializer.ErrorResponse
// @Router /users/register [post]
func (h *Handler) ApiRegisterUser(ctx *gin.Context) {
	var body serializer.RegisterUserJSON
	if err := ctx.BindJSON(&body); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	u, err := h.Repository.CreateUserWithPassword(strings.TrimSpace(body.Name), strings.TrimSpace(body.Email), body.Password)
	if err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.Header("Location", fmt.Sprintf("/api/users/%d", u.ID))
	ctx.JSON(http.StatusCreated, serializer.UserToJSON(u))
}

// ApiLogin godoc
// @Summary Логин (JWT)
// @Tags users
// @Accept json
// @Produce json
// @Param body body serializer.LoginJSON true "Учетные данные"
// @Success 200 {object} serializer.LoginResponseJSON
// @Failure 400 {object} serializer.ErrorResponse
// @Failure 401 {object} serializer.ErrorResponse
// @Router /users/login [post]
func (h *Handler) ApiLogin(ctx *gin.Context) {
	var body serializer.LoginJSON
	if err := ctx.BindJSON(&body); err != nil {
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	user, token, err := h.Repository.SignIn(body.Email, body.Password)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, repository.ErrNotAllowed) {
			h.apiError(ctx, http.StatusUnauthorized, err)
			return
		}
		h.apiError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.SetCookie("access_token", token, 3600, "/", "", false, true)
	ctx.JSON(http.StatusOK, serializer.LoginResponseJSON{Token: token, Role: user.Role})
}

// ApiLogout godoc
// @Summary Logout (blacklist token)
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} serializer.ErrorResponse
// @Router /users/logout [post]
func (h *Handler) ApiLogout(ctx *gin.Context) {
	user, ok := currentUser(ctx)
	if !ok {
		h.apiError(ctx, http.StatusUnauthorized, repository.ErrNotAllowed)
		return
	}
	tokenString := extractToken(ctx)
	if tokenString != "" {
		ttlVal, _ := ctx.Get("auth_token_ttl")
		ttl, _ := ttlVal.(time.Duration)
		if err := h.Repository.AddTokenToBlacklist(ctx.Request.Context(), tokenString, ttl, user.ID); err != nil {
			h.apiError(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	ctx.SetCookie("access_token", "", -1, "/", "", false, true)
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
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
