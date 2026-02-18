package handler

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"sql-index-simulator/internal/app/repository"
)

// Handler wraps dependencies for HTTP handlers.
type Handler struct {
	repo *repository.Repository
}

// NewHandler creates new Handler.
func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

// GetServicesPage handles the main page with list and search.
func (h *Handler) GetServicesPage(c *gin.Context) {
	query := c.Query("query")

	var (
		services []repository.Service
		err      error
	)

	if query != "" {
		services, err = h.repo.GetServicesByName(query)
	} else {
		services, err = h.repo.GetAllServices()
	}

	if err != nil {
		logrus.WithError(err).Error("failed to get services")
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Time":     time.Now().Format("02.01.2006 15:04:05"),
		"Services": services,
		"Query":    query,
	})
}

// CalculationResult holds simulation results.
type CalculationResult struct {
	TableSize       int
	ExecutionTimeMs float64
	MemoryUsageMB   float64
	IndexType       string
}

// simulateCalculation performs a very simple pseudo-calculation.
func simulateCalculation(s repository.Service, tableSize int) CalculationResult {
	if tableSize <= 0 {
		tableSize = 1
	}

	var execTime float64
	switch s.Name {
	case "B-tree":
		// O(log N)
		execTime = s.BaseSpeed * math.Log2(float64(tableSize)+1) * 10
	case "Hash":
		// O(1) with some overhead
		execTime = s.BaseSpeed * 5
	case "Bitmap":
		// O(N) scaled down
		execTime = s.BaseSpeed * math.Sqrt(float64(tableSize))
	default:
		execTime = s.BaseSpeed * math.Log(float64(tableSize)+1)
	}

	var memory float64
	switch s.Name {
	case "B-tree":
		memory = s.BaseMemory * math.Log2(float64(tableSize)+1)
	case "Hash":
		memory = s.BaseMemory * float64(tableSize) * 0.0001
	case "Bitmap":
		memory = s.BaseMemory * float64(tableSize) * 0.00005
	default:
		memory = s.BaseMemory
	}

	return CalculationResult{
		TableSize:       tableSize,
		ExecutionTimeMs: math.Round(execTime*100) / 100,
		MemoryUsageMB:   math.Round(memory*100) / 100,
		IndexType:       s.Name,
	}
}

// GetServicePage handles detail page and calculation form.
func (h *Handler) GetServicePage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.WithError(err).WithField("id", idStr).Warn("invalid service id")
		c.String(http.StatusBadRequest, "invalid service id")
		return
	}

	service, err := h.repo.GetServiceByID(id)
	if err != nil {
		if err == repository.ErrNotFound {
			c.String(http.StatusNotFound, "service not found")
			return
		}
		logrus.WithError(err).WithField("id", id).Error("failed to get service by id")
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	tableSizeStr := c.Query("tableSize")
	var (
		tableSize int
		calcRes   *CalculationResult
	)

	if tableSizeStr != "" {
		tableSize, err = strconv.Atoi(tableSizeStr)
		if err != nil || tableSize <= 0 {
			logrus.WithError(err).WithField("tableSize", tableSizeStr).Warn("invalid table size")
			c.HTML(http.StatusBadRequest, "service.html", gin.H{
				"Service": service,
				"Error":   "Некорректный размер таблицы. Введите положительное целое число.",
			})
			return
		}
		res := simulateCalculation(service, tableSize)
		calcRes = &res
	}

	c.HTML(http.StatusOK, "service.html", gin.H{
		"Service": service,
		"Result":  calcRes,
	})
}


