package repository

import (
	"errors"
	"strings"
)

// Service represents a simulated index type "service".
type Service struct {
	ID          int
	Name        string
	Description string
	BaseSpeed   float64
	BaseMemory  float64
}

// Repository provides access to mock data.
type Repository struct {
	services []Service
}

// NewRepository initializes repository with mock data.
func NewRepository() *Repository {
	return &Repository{
		services: []Service{
			{
				ID:          1,
				Name:        "B-tree",
				Description: "Сбалансированное дерево для эффективного поиска по диапазонам и точным значениям.",
				BaseSpeed:   1.0,
				BaseMemory:  1.2,
			},
			{
				ID:          2,
				Name:        "Hash",
				Description: "Хеш-индекс для быстрых точечных запросов по равенству.",
				BaseSpeed:   0.7,
				BaseMemory:  1.5,
			},
			{
				ID:          3,
				Name:        "Bitmap",
				Description: "Битмап-индекс для аналитических запросов и низкой кардинальности.",
				BaseSpeed:   0.9,
				BaseMemory:  0.8,
			},
		},
	}
}

// GetAllServices returns all services.
func (r *Repository) GetAllServices() ([]Service, error) {
	return r.services, nil
}

// GetServicesByName filters services by name (case-insensitive substring).
func (r *Repository) GetServicesByName(name string) ([]Service, error) {
	if name == "" {
		return r.GetAllServices()
	}

	lower := strings.ToLower(name)
	result := make([]Service, 0)
	for _, s := range r.services {
		if strings.Contains(strings.ToLower(s.Name), lower) {
			result = append(result, s)
		}
	}
	return result, nil
}

// ErrNotFound is returned when a service is not found.
var ErrNotFound = errors.New("service not found")

// GetServiceByID returns service by ID.
func (r *Repository) GetServiceByID(id int) (Service, error) {
	for _, s := range r.services {
		if s.ID == id {
			return s, nil
		}
	}
	return Service{}, ErrNotFound
}


