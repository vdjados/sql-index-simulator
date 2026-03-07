package repository

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Статусы заявок
const (
	StatusDraft    = "draft"
	StatusDeleted  = "deleted"
	StatusFormed   = "formed"
	StatusFinished = "completed"
	StatusRejected = "rejected"
)

type Repository struct {
	db *gorm.DB
}

// Модели БД

type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:255;not null"`
	Email string `gorm:"size:255;uniqueIndex"`
	Role  string `gorm:"size:32;not null;default:'user'"`
}

// Service описывает тип индекса в симуляторе и хранится в таблице services.
type Service struct {
	ID          string `gorm:"primaryKey;size:64"`
	Name        string `gorm:"size:255;not null"`
	TableSize   string `gorm:"size:64;not null"`
	Speed       string `gorm:"size:64;not null"`
	Description string `gorm:"type:text;not null"`
	ImageKey    string `gorm:"size:255"`
	GifKey      string `gorm:"size:255"`
	Status      string `gorm:"size:32;not null;default:'active'"` // active / deleted
}

// Request описывает заявку (симуляцию запроса) и её сводный результат.
type Request struct {
	ID uint `gorm:"primaryKey"`

	Status string `gorm:"size:32;not null"`

	CreatedAt   time.Time `gorm:"not null"`
	CreatedByID uint      `gorm:"not null"`

	FormedAt   *time.Time
	FinishedAt *time.Time

	ModeratorID *uint

	Selectivity  float64 `gorm:"not null"`
	ResultTime   string  `gorm:"size:64"`
	ResultMemory string  `gorm:"size:64"`

	Services []RequestService `gorm:"foreignKey:RequestID"`
}

// RequestService описывает связь m-n между заявкой и услугой.
// Составной уникальный ключ (request_id, service_id). При завершении заявки заполняется CalculatedTimeMs.
type RequestService struct {
	RequestID uint   `gorm:"primaryKey"`
	ServiceID string `gorm:"primaryKey;size:64"`

	ServiceName string  `gorm:"size:255;not null"`
	TableSize   string  `gorm:"size:64;not null"`
	Selectivity float64 `gorm:"not null"`
	ImageKey    string  `gorm:"size:255"`

	Quantity int   `gorm:"not null;default:1"`
	Position int   `gorm:"not null;default:1"`
	IsMain   bool  `gorm:"not null;default:false"`
	CalculatedTimeMs float64 `gorm:"type:numeric(10,2)"` // считается при завершении заявки

	Service *Service `gorm:"foreignKey:ServiceID"` // для расчёта по формуле (скорость индекса)
}

// NewRepository инициализирует подключение к БД и выполняет миграции.
func NewRepository(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&User{}, &Service{}, &Request{}, &RequestService{}); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

// parseSpeedMs извлекает число из строки вида "0.3ms" или "1.8ms".
func parseSpeedMs(s string) float64 {
	s = strings.TrimSpace(strings.TrimSuffix(strings.ToLower(s), "ms"))
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// calculateResultTime считает время по формуле: Σ (quantity × t_i × (1 + selectivity)).
func (r *Repository) calculateResultTimeAndMemory(req *Request) (timeMs float64, memoryKB float64) {
	for i := range req.Services {
		rs := &req.Services[i]
		var speedMs float64
		if rs.Service != nil {
			speedMs = parseSpeedMs(rs.Service.Speed)
		}
		timeMs += float64(rs.Quantity) * speedMs * (1 + rs.Selectivity)
	}
	// Память: базовые 64 KB + примерно по 16 KB на каждую позицию в заявке
	memoryKB = 64 + float64(len(req.Services))*16
	return timeMs, memoryKB
}

// GetServices возвращает список индексов с фильтрацией по имени или размеру таблицы.
// Фильтр применяется по подстроке без учёта регистра, учитываются только активные услуги.
func (r *Repository) GetServices(filter string) ([]Service, error) {
	var services []Service
	query := r.db.Where("status = ?", "active")

	if filter != "" {
		lower := strings.ToLower(filter)
		like := "%" + lower + "%"
		query = query.Where(
			r.db.Where("LOWER(name) LIKE ?", like).
				Or("LOWER(table_size) LIKE ?", like),
		)
	}

	if err := query.Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

// GetService возвращает одну услугу по её ID.
func (r *Repository) GetService(id string) (Service, error) {
	var s Service
	if err := r.db.First(&s, "id = ? AND status = ?", id, "active").Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Service{}, fmt.Errorf("индекс с ID %s не найден", id)
		}
		return Service{}, err
	}
	return s, nil
}

// GetRequest возвращает одну заявку по ID вместе с услугами. Время и память считаются по формуле, если ещё не сохранены.
func (r *Repository) GetRequest(id uint) (Request, error) {
	var req Request
	if err := r.db.Preload("Services.Service").First(&req, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Request{}, fmt.Errorf("запрос с ID %d не найден", id)
		}
		return Request{}, err
	}

	if req.Status == StatusDeleted {
		return Request{}, fmt.Errorf("запрос с ID %d удалён", id)
	}

	// Расчёт по формуле, если в заявке есть услуги и результат ещё не сохранён
	if len(req.Services) > 0 && req.ResultTime == "" {
		timeMs, memKB := r.calculateResultTimeAndMemory(&req)
		req.ResultTime = fmt.Sprintf("%.2fms", timeMs)
		req.ResultMemory = fmt.Sprintf("%.0fKB", memKB)
	}
	if req.ResultTime == "" {
		req.ResultTime = "—"
	}
	if req.ResultMemory == "" {
		req.ResultMemory = "—"
	}
	return req, nil
}

// GetCurrentRequest ищет текущую заявку пользователя в статусе черновика.
func (r *Repository) GetCurrentRequest(userID uint) (*Request, error) {
	var req Request
	err := r.db.Preload("Services.Service").First(&req, "created_by_id = ? AND status = ?", userID, StatusDraft).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if len(req.Services) > 0 && req.ResultTime == "" {
		timeMs, memKB := r.calculateResultTimeAndMemory(&req)
		req.ResultTime = fmt.Sprintf("%.2fms", timeMs)
		req.ResultMemory = fmt.Sprintf("%.0fKB", memKB)
	}
	return &req, nil
}

// AddServiceToDraft добавляет услугу в текущую заявку-проект пользователя.
// Если черновика нет, он создаётся. Возвращает итоговую заявку.
func (r *Repository) AddServiceToDraft(userID uint, serviceID string) (*Request, error) {
	var service Service
	if err := r.db.First(&service, "id = ? AND status = ?", serviceID, "active").Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("индекс с ID %s не найден", serviceID)
		}
		return nil, err
	}

	var req Request
	if err := r.db.Where("created_by_id = ? AND status = ?", userID, StatusDraft).First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			req = Request{
				Status:       StatusDraft,
				CreatedAt:    time.Now(),
				CreatedByID:  userID,
				Selectivity:  0.05, // базовое значение, можно править через форму
				ResultTime:   "",
				ResultMemory: "",
			}
			if err := r.db.Create(&req).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var existing RequestService
	err := r.db.First(&existing, "request_id = ? AND service_id = ?", req.ID, service.ID).Error
	if err == nil {
		existing.Quantity++
		if err := r.db.Save(&existing).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		var count int64
		if err := r.db.Model(&RequestService{}).Where("request_id = ?", req.ID).Count(&count).Error; err != nil {
			return nil, err
		}

		rs := RequestService{
			RequestID:   req.ID,
			ServiceID:   service.ID,
			ServiceName: service.Name,
			TableSize:   service.TableSize,
			Selectivity: req.Selectivity,
			ImageKey:    service.ImageKey,
			Quantity:    1,
			Position:    int(count) + 1,
			IsMain:      count == 0,
		}

		if err := r.db.Create(&rs).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	if err := r.db.Preload("Services").First(&req, "id = ?", req.ID).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

// DeleteRequestLogical выполняет логическое удаление заявки через raw SQL UPDATE без использования ORM.
func (r *Repository) DeleteRequestLogical(userID uint, requestID uint) error {
	sql := "UPDATE requests SET status = $1 WHERE id = $2 AND created_by_id = $3"
	result := r.db.Exec(sql, StatusDeleted, requestID, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("заявка не найдена или недоступна для удаления")
	}
	return nil
}

// CompleteRequest переводит заявку в статус «завершена», считает по формуле время и память и сохраняет в БД.
// В таблицу request_services записывается рассчитанное время по каждой позиции (CalculatedTimeMs).
func (r *Repository) CompleteRequest(userID uint, requestID uint) error {
	var req Request
	if err := r.db.Preload("Services.Service").First(&req, "id = ? AND created_by_id = ?", requestID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("заявка не найдена")
		}
		return err
	}
	if req.Status != StatusDraft {
		return fmt.Errorf("завершить можно только заявку в статусе черновик")
	}
	if len(req.Services) == 0 {
		return fmt.Errorf("в заявке нет услуг")
	}

	timeMs, memKB := r.calculateResultTimeAndMemory(&req)
	now := time.Now()

	// Заполняем CalculatedTimeMs по каждой позиции м-м
	for i := range req.Services {
		rs := &req.Services[i]
		var speedMs float64
		if rs.Service != nil {
			speedMs = parseSpeedMs(rs.Service.Speed)
		}
		rs.CalculatedTimeMs = float64(rs.Quantity) * speedMs * (1 + rs.Selectivity)
		if err := r.db.Model(rs).Update("calculated_time_ms", rs.CalculatedTimeMs).Error; err != nil {
			return err
		}
	}

	req.Status = StatusFinished
	req.ResultTime = fmt.Sprintf("%.2fms", timeMs)
	req.ResultMemory = fmt.Sprintf("%.0fKB", memKB)
	req.FinishedAt = &now
	return r.db.Save(&req).Error
}
