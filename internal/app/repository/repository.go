package repository

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	minioClient "web_backend/internal/app/minioClient"
)

// Статусы sql_query
const (
	StatusDraft    = "draft"
	StatusDeleted  = "deleted"
	StatusFormed   = "formed"
	StatusFinished = "completed"
	StatusRejected = "rejected"
)

type Repository struct {
	db *gorm.DB
	mc *minio.Client
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

// Request описывает sql_query (симуляцию запроса) и её сводный результат.
type Request struct {
	ID uint `gorm:"primaryKey"`

	Status string `gorm:"size:32;not null"`

	CreatedAt   time.Time `gorm:"not null"`
	CreatedByID uint      `gorm:"not null"`

	FormedAt   *time.Time
	FinishedAt *time.Time

	ModeratorID *uint

	QueryDescription string  `gorm:"type:text"` // описание запроса текстом (поле Заявка)
	Selectivity      float64 `gorm:"not null"`  // селективность в запросе (м-м)
	ResultTime       string  `gorm:"size:64"`   // время запроса
	ResultMemory     string  `gorm:"size:64"`   // память запроса

	Services []RequestService `gorm:"foreignKey:RequestID"`
}

// RequestService описывает связь м-м между sql_query и услугой (таблица+индекс).
// Составной уникальный ключ (request_id, service_id). При завершении заявки заполняется CalculatedTimeMs.
type RequestService struct {
	RequestID uint   `gorm:"primaryKey"`
	ServiceID string `gorm:"primaryKey;size:64"`

	ServiceName string  `gorm:"size:255;not null"`
	TableSize   string  `gorm:"size:64;not null"`
	Selectivity float64 `gorm:"not null"`
	ImageKey    string  `gorm:"size:255"`

	Quantity         int     `gorm:"not null;default:1"`
	Position         int     `gorm:"not null;default:1"`
	IsMain           bool    `gorm:"not null;default:false"`
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

	mc, err := minioClient.InitMinio()
	if err != nil {
		return nil, err
	}

	return &Repository{db: db, mc: mc}, nil
}

const creatorUserID uint = 1

func CreatorUserID() uint {
	return creatorUserID
}

const moderatorUserID uint = 2

func ModeratorUserID() uint {
	return moderatorUserID
}

var (
	ErrNotFound   = errors.New("not found")
	ErrNotAllowed = errors.New("not allowed")
	ErrNoDraft    = errors.New("no draft")
	ErrValidation = errors.New("validation")
)

func (r *Repository) CreateUser(u User) (User, error) {
	u.ID = 0
	if u.Role == "" {
		u.Role = "user"
	}
	if u.Name == "" || u.Email == "" {
		return User{}, ErrValidation
	}
	if err := r.db.Create(&u).Error; err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *Repository) getDraftOrCreate(userID uint) (*Request, error) {
	var req Request
	err := r.db.Preload("Services.Service").First(&req, "created_by_id = ? AND status = ?", userID, StatusDraft).Error
	if err == nil {
		return &req, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	req = Request{
		Status:           StatusDraft,
		CreatedAt:        time.Now(),
		CreatedByID:      userID,
		Selectivity:      0.05,
		ResultTime:       "",
		ResultMemory:     "",
		QueryDescription: "",
	}
	if err := r.db.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Repository) ApiGetSqlQuery(id uint) (Request, error) {
	var req Request
	if err := r.db.Preload("Services.Service").First(&req, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Request{}, ErrNotFound
		}
		return Request{}, err
	}
	if req.Status == StatusDeleted {
		return Request{}, ErrNotFound
	}
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

func (r *Repository) ApiEditSqlQuery(userID uint, id uint, queryDescription *string, selectivity *float64) (Request, error) {
	req, err := r.ApiGetSqlQuery(id)
	if err != nil {
		return Request{}, err
	}
	if req.CreatedByID != userID {
		return Request{}, ErrNotAllowed
	}
	if req.Status != StatusDraft {
		return Request{}, ErrNotAllowed
	}
	if queryDescription != nil {
		req.QueryDescription = strings.TrimSpace(*queryDescription)
	}
	if selectivity != nil {
		if *selectivity < 0 || *selectivity > 1 {
			return Request{}, ErrValidation
		}
		req.Selectivity = *selectivity
	}
	if err := r.db.Save(&req).Error; err != nil {
		return Request{}, err
	}
	// синхронизируем селективность в m-m, если она используется в позициях
	if selectivity != nil {
		_ = r.db.Model(&RequestService{}).Where("request_id = ?", req.ID).Update("selectivity", req.Selectivity).Error
	}
	return req, nil
}

func (r *Repository) ApiDeleteSqlQuery(userID uint, id uint) error {
	req, err := r.ApiGetSqlQuery(id)
	if err != nil {
		return err
	}
	if req.CreatedByID != userID {
		return ErrNotAllowed
	}
	if req.Status != StatusDraft {
		return ErrNotAllowed
	}
	sql := "UPDATE requests SET status = $1 WHERE id = $2 AND created_by_id = $3"
	res := r.db.Exec(sql, StatusDeleted, id, userID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ApiAddToDraft(userID uint, serviceID string) (RequestService, error) {
	draft, err := r.getDraftOrCreate(userID)
	if err != nil {
		return RequestService{}, err
	}
	if draft.Status != StatusDraft {
		return RequestService{}, ErrNotAllowed
	}

	var svc Service
	if err := r.db.First(&svc, "id = ? AND status = ?", serviceID, "active").Error; err != nil {
		return RequestService{}, ErrNotFound
	}

	var item RequestService
	err = r.db.First(&item, "request_id = ? AND service_id = ?", draft.ID, svc.ID).Error
	if err == nil {
		item.Quantity++
		if err := r.db.Save(&item).Error; err != nil {
			return RequestService{}, err
		}
		return item, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return RequestService{}, err
	}

	var count int64
	if err := r.db.Model(&RequestService{}).Where("request_id = ?", draft.ID).Count(&count).Error; err != nil {
		return RequestService{}, err
	}
	item = RequestService{
		RequestID:   draft.ID,
		ServiceID:   svc.ID,
		ServiceName: svc.Name,
		TableSize:   svc.TableSize,
		Selectivity: draft.Selectivity,
		ImageKey:    svc.ImageKey,
		Quantity:    1,
		Position:    int(count) + 1,
		IsMain:      count == 0,
	}
	if err := r.db.Create(&item).Error; err != nil {
		return RequestService{}, err
	}
	return item, nil
}

func (r *Repository) ApiEditItem(userID uint, sqlQueryID uint, serviceID string, quantity *int, position *int, selectivity *float64) (RequestService, error) {
	req, err := r.ApiGetSqlQuery(sqlQueryID)
	if err != nil {
		return RequestService{}, err
	}
	if req.CreatedByID != userID || req.Status != StatusDraft {
		return RequestService{}, ErrNotAllowed
	}
	var item RequestService
	if err := r.db.First(&item, "request_id = ? AND service_id = ?", sqlQueryID, serviceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return RequestService{}, ErrNotFound
		}
		return RequestService{}, err
	}
	if quantity != nil {
		if *quantity < 1 {
			return RequestService{}, ErrValidation
		}
		item.Quantity = *quantity
	}
	if position != nil {
		if *position < 1 {
			return RequestService{}, ErrValidation
		}
		item.Position = *position
	}
	if selectivity != nil {
		if *selectivity < 0 || *selectivity > 1 {
			return RequestService{}, ErrValidation
		}
		item.Selectivity = *selectivity
	}
	if err := r.db.Save(&item).Error; err != nil {
		return RequestService{}, err
	}
	return item, nil
}

func (r *Repository) ApiDeleteItem(userID uint, sqlQueryID uint, serviceID string) error {
	req, err := r.ApiGetSqlQuery(sqlQueryID)
	if err != nil {
		return err
	}
	if req.CreatedByID != userID || req.Status != StatusDraft {
		return ErrNotAllowed
	}
	res := r.db.Delete(&RequestService{}, "request_id = ? AND service_id = ?", sqlQueryID, serviceID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ApiFormSqlQuery(userID uint, id uint) (Request, error) {
	req, err := r.ApiGetSqlQuery(id)
	if err != nil {
		return Request{}, err
	}
	if req.CreatedByID != userID {
		return Request{}, ErrNotAllowed
	}
	if req.Status != StatusDraft {
		return Request{}, ErrNotAllowed
	}
	if strings.TrimSpace(req.QueryDescription) == "" {
		return Request{}, ErrValidation
	}
	if err := r.db.Preload("Services.Service").First(&req, "id = ?", id).Error; err != nil {
		return Request{}, err
	}
	if len(req.Services) == 0 {
		return Request{}, ErrValidation
	}

	// рассчитать результат и записать в БД + записать результат по позициям
	timeMs, memKB := r.calculateResultTimeAndMemory(&req)
	for i := range req.Services {
		rs := &req.Services[i]
		var speedMs float64
		if rs.Service != nil {
			speedMs = parseSpeedMs(rs.Service.Speed)
		}
		rs.CalculatedTimeMs = float64(rs.Quantity) * speedMs * (1 + rs.Selectivity)
		if err := r.db.Model(rs).Update("calculated_time_ms", rs.CalculatedTimeMs).Error; err != nil {
			return Request{}, err
		}
	}

	now := time.Now()
	req.Status = StatusFormed
	req.FormedAt = &now
	req.ResultTime = fmt.Sprintf("%.2fms", timeMs)
	req.ResultMemory = fmt.Sprintf("%.0fKB", memKB)
	if err := r.db.Save(&req).Error; err != nil {
		return Request{}, err
	}
	return req, nil
}

func (r *Repository) ApiFinishSqlQuery(id uint, status string) (Request, error) {
	req, err := r.ApiGetSqlQuery(id)
	if err != nil {
		return Request{}, err
	}
	if req.Status != StatusFormed {
		return Request{}, ErrNotAllowed
	}
	if status != StatusFinished && status != StatusRejected {
		return Request{}, ErrValidation
	}
	now := time.Now()
	mid := ModeratorUserID()
	req.ModeratorID = &mid
	req.FinishedAt = &now
	req.Status = status
	if err := r.db.Save(&req).Error; err != nil {
		return Request{}, err
	}
	return req, nil
}

func (r *Repository) GetModeratorAndCreatorLogin(q Request) (creatorLogin, moderatorLogin string, err error) {
	var creator User
	if err := r.db.First(&creator, "id = ?", q.CreatedByID).Error; err == nil {
		if creator.Email != "" {
			creatorLogin = creator.Email
		} else {
			creatorLogin = creator.Name
		}
	}
	if q.ModeratorID != nil {
		var moderator User
		if err := r.db.First(&moderator, "id = ?", *q.ModeratorID).Error; err == nil {
			if moderator.Email != "" {
				moderatorLogin = moderator.Email
			} else {
				moderatorLogin = moderator.Name
			}
		}
	}
	return creatorLogin, moderatorLogin, nil
}

func (r *Repository) AddServiceMedia(ctx context.Context, serviceID string, imageFile, videoFile *multipart.FileHeader) (Service, error) {
	var s Service
	if err := r.db.First(&s, "id = ? AND status = ?", serviceID, "active").Error; err != nil {
		return Service{}, ErrNotFound
	}

	bucket := minioClient.Bucket()
	if imageFile != nil {
		key, err := minioClient.UploadFile(ctx, r.mc, bucket, "service_img_"+serviceID, imageFile)
		if err != nil {
			return Service{}, err
		}
		s.ImageKey = key
	}
	if videoFile != nil {
		key, err := minioClient.UploadFile(ctx, r.mc, bucket, "service_vid_"+serviceID, videoFile)
		if err != nil {
			return Service{}, err
		}
		s.GifKey = key
	}
	if err := r.db.Save(&s).Error; err != nil {
		return Service{}, err
	}
	return s, nil
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

func (r *Repository) CreateService(s Service) error {
	// не даём создавать deleted
	s.Status = "active"
	return r.db.Create(&s).Error
}

// ApiListSqlQueries возвращает список sql_query для API: исключаем draft и deleted.
// Фильтруем по статусу и диапазону даты формирования.
func (r *Repository) ApiListSqlQueries(from, to time.Time, status string) ([]Request, error) {
	var qs []Request
	q := r.db.Model(&Request{}).Where("status <> ? AND status <> ?", StatusDraft, StatusDeleted)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if !from.IsZero() {
		q = q.Where("formed_at >= ?", from)
	}
	if !to.IsZero() {
		// включительно по дате: +1 день
		q = q.Where("formed_at < ?", to.Add(24*time.Hour))
	}
	if err := q.Order("id desc").Find(&qs).Error; err != nil {
		return nil, err
	}
	return qs, nil
}

// GetCompletedItemCount считает количество m-m записей с непустым рассчитанным полем.
func (r *Repository) GetCompletedItemCount(sqlQueryID uint) int {
	var count int64
	_ = r.db.Model(&RequestService{}).
		Where("request_id = ? AND calculated_time_ms IS NOT NULL AND calculated_time_ms <> 0", sqlQueryID).
		Count(&count).Error
	return int(count)
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

// AddServiceToDraft добавляет услугу (таблица+индекс) в текущий sql_query (черновик) пользователя.
// Если черновика нет, он создаётся. Возвращает итоговый sql_query.
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
		return fmt.Errorf("sql_query не найден или недоступен для удаления")
	}
	return nil
}

// CompleteRequest переводит заявку в статус «завершена», считает по формуле время и память и сохраняет в БД.
// В таблицу request_services записывается рассчитанное время по каждой позиции (CalculatedTimeMs).
func (r *Repository) CompleteRequest(userID uint, requestID uint) error {
	var req Request
	if err := r.db.Preload("Services.Service").First(&req, "id = ? AND created_by_id = ?", requestID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("sql_query не найден")
		}
		return err
	}
	if req.Status != StatusDraft {
		return fmt.Errorf("завершить можно только sql_query в статусе черновик")
	}
	if len(req.Services) == 0 {
		return fmt.Errorf("в sql_query нет услуг")
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
