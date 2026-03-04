package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

// Service описывает тип индекса в симуляторе
type Service struct {
	ID          string
	Name        string
	TableSize   string
	Speed       string
	Description string
	ImageKey    string
	GifKey      string
}

// Request описывает запрос (запрос) и его готовый результат
type Request struct {
	ID           string
	Selectivity  float64
	ResultTime   string
	ResultMemory string
	Services     []RequestService
}

// RequestService связывает запрос с конкретным индексом
type RequestService struct {
	ServiceID   string
	ServiceName string
	TableSize   string
	Selectivity float64
	ImageKey    string
}

func (r *Repository) getAllServices() []Service {
	return []Service{
		{
			ID:        "btree-10k",
			Name:      "B-tree",
			TableSize: "10k rows",
			Speed:     "0.3ms",
			Description: "Классический сбалансированный B-tree индекс, оптимальный для диапазонных и точечных запросов по упорядоченным ключам. Хорошо масштабируется и обеспечивает логарифмическую сложность поиска.",
			ImageKey: "btree.png",
			GifKey:   "btree.gif",
		},
		{
			ID:        "btree-1m",
			Name:      "B-tree (large)",
			TableSize: "1M rows",
			Speed:     "1.8ms",
			Description: "B-tree индекс на крупной таблице, демонстрирует стабильное время отклика даже при миллионах строк благодаря иерархической структуре страниц.",
			ImageKey: "btree.png",
			GifKey:   "btree.gif",
		},
		{
			ID:        "hash-100k",
			Name:      "Hash",
			TableSize: "100k rows",
			Speed:     "0.5ms",
			Description: "Hash индекс, оптимизированный для точечных запросов по равенству. Не поддерживает диапазонные операции, но показывает высокую производительность при выборке по ключу.",
			ImageKey: "hash.png",
			GifKey:   "hash.gif",
		},
		{
			ID:        "bitmap-1m",
			Name:      "Bitmap",
			TableSize: "1M rows",
			Speed:     "2.1ms",
			Description: "Bitmap индекс для низкоселективных столбцов. Эффективен для сложных логических комбинаций условий и аналитических запросов.",
			ImageKey: "bitmap.png",
			GifKey:   "bitmap.gif",
		},
		{
			ID:        "gist-500k",
			Name:      "GiST",
			TableSize: "500k rows",
			Speed:     "1.2ms",
			Description: "GiST (Generalized Search Tree) индекс, используемый для сложных типов данных: геометрии, диапазоны, полнотекстовый поиск. Поддерживает настраиваемые операторы соответствия.",
			ImageKey: "gist.png",
			GifKey:   "gist.gif",
		},
		{
			ID:        "gin-500k",
			Name:      "GIN",
			TableSize: "500k rows",
			Speed:     "0.9ms",
			Description: "GIN (Generalized Inverted Index) индекс, оптимальный для массивов и полнотекстового поиска. Обеспечивает быстрый поиск по множеству значений в одной записи.",
			ImageKey: "gin.png",
			GifKey:   "gin.gif",
		},
	}
}

// GetServices возвращает список индексов с серверной фильтрацией по имени или размеру таблицы.
// Фильтр применяется по подстроке без учёта регистра.
func (r *Repository) GetServices(filter string) ([]Service, error) {
	services := r.getAllServices()
	if filter == "" {
		return services, nil
	}

	lower := strings.ToLower(filter)
	var result []Service
	for _, s := range services {
		if strings.Contains(strings.ToLower(s.Name), lower) ||
			strings.Contains(strings.ToLower(s.TableSize), lower) {
			result = append(result, s)
		}
	}

	if len(result) == 0 {
		return []Service{}, nil
	}
	return result, nil
}

// GetService возвращает один индекс по его ID.
func (r *Repository) GetService(id string) (Service, error) {
	for _, s := range r.getAllServices() {
		if s.ID == id {
			return s, nil
		}
	}
	return Service{}, fmt.Errorf("индекс с ID %s не найден", id)
}

// getAllRequests возвращает заранее подготовленные запросы (запросы) с готовыми результатами.
func (r *Repository) getAllRequests() []Request {
	services := r.getAllServices()
	serviceByID := make(map[string]Service, len(services))
	for _, s := range services {
		serviceByID[s.ID] = s
	}

	req1 := Request{
		ID:           "1",
		Selectivity:  0.05,
		ResultTime:   "0.7ms",
		ResultMemory: "64KB",
		Services: []RequestService{
			{
				ServiceID:   "btree-10k",
				ServiceName: serviceByID["btree-10k"].Name,
				TableSize:   serviceByID["btree-10k"].TableSize,
				Selectivity: 0.05,
				ImageKey:    serviceByID["btree-10k"].ImageKey,
			},
			{
				ServiceID:   "hash-100k",
				ServiceName: serviceByID["hash-100k"].Name,
				TableSize:   serviceByID["hash-100k"].TableSize,
				Selectivity: 0.05,
				ImageKey:    serviceByID["hash-100k"].ImageKey,
			},
			{
				ServiceID:   "gin-500k",
				ServiceName: serviceByID["gin-500k"].Name,
				TableSize:   serviceByID["gin-500k"].TableSize,
				Selectivity: 0.05,
				ImageKey:    serviceByID["gin-500k"].ImageKey,
			},
		},
	}

	req2 := Request{
		ID:           "2",
		Selectivity:  0.5,
		ResultTime:   "3.4ms",
		ResultMemory: "128KB",
		Services: []RequestService{
			{
				ServiceID:   "btree-1m",
				ServiceName: serviceByID["btree-1m"].Name,
				TableSize:   serviceByID["btree-1m"].TableSize,
				Selectivity: 0.5,
				ImageKey:    serviceByID["btree-1m"].ImageKey,
			},
			{
				ServiceID:   "bitmap-1m",
				ServiceName: serviceByID["bitmap-1m"].Name,
				TableSize:   serviceByID["bitmap-1m"].TableSize,
				Selectivity: 0.5,
				ImageKey:    serviceByID["bitmap-1m"].ImageKey,
			},
			{
				ServiceID:   "gist-500k",
				ServiceName: serviceByID["gist-500k"].Name,
				TableSize:   serviceByID["gist-500k"].TableSize,
				Selectivity: 0.5,
				ImageKey:    serviceByID["gist-500k"].ImageKey,
			},
		},
	}

	return []Request{req1, req2}
}

// GetRequests возвращает все запросы.
func (r *Repository) GetRequests() ([]Request, error) {
	return r.getAllRequests(), nil
}

// GetRequest возвращает один запрос по ID.
func (r *Repository) GetRequest(id string) (Request, error) {
	for _, req := range r.getAllRequests() {
		if req.ID == id {
			return req, nil
		}
	}
	return Request{}, fmt.Errorf("запрос с ID %s не найден", id)
}

// GetCurrentRequest возвращает первый запрос для отображения в "корзине" на главной странице.
func (r *Repository) GetCurrentRequest() (*Request, error) {
	requests := r.getAllRequests()
	if len(requests) == 0 {
		return nil, fmt.Errorf("список запросов пуст")
	}
	return &requests[0], nil
}
