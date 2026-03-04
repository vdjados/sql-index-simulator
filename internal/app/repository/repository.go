package repository

import (
	"errors"
	"fmt"
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
// Используется составной ключ (request_id, service_id) и дополнительные поля.
type RequestService struct {
	RequestID uint   `gorm:"primaryKey"`
	ServiceID string `gorm:"primaryKey;size:64"`

	// Дублирующие поля по предметной области для удобства отображения
	ServiceName string  `gorm:"size:255;not null"`
	TableSize   string  `gorm:"size:64;not null"`
	Selectivity float64 `gorm:"not null"`
	ImageKey    string  `gorm:"size:255"`

	Quantity int  `gorm:"not null;default:1"`
	Position int  `gorm:"not null;default:1"`
	IsMain   bool `gorm:"not null;default:false"`
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

// GetRequest возвращает одну заявку по ID вместе с её услугами.
func (r *Repository) GetRequest(id uint) (Request, error) {
	var req Request
	if err := r.db.Preload("Services").First(&req, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Request{}, fmt.Errorf("запрос с ID %d не найден", id)
		}
		return Request{}, err
	}

	// Логически удалённые заявки не должны быть доступны.
	if req.Status == StatusDeleted {
		return Request{}, fmt.Errorf("запрос с ID %d удалён", id)
	}

	return req, nil
}

// GetCurrentRequest ищет текущую заявку пользователя в статусе черновика.
func (r *Repository) GetCurrentRequest(userID uint) (*Request, error) {
	var req Request
	err := r.db.Preload("Services").First(&req, "created_by_id = ? AND status = ?", userID, StatusDraft).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
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
