package serializer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"web_backend/internal/app/repository"
)

type SqlQueryJSON struct {
	ID             uint       `json:"id"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	CreatedByID    uint       `json:"created_by_id"`
	FormedAt       *time.Time `json:"formed_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	ModeratorID    *uint      `json:"moderator_id"`
	CreatorLogin   string     `json:"creator_login"`
	ModeratorLogin *string    `json:"moderator_login"`
	QueryText      string     `json:"query_text"`
	Theme          string     `json:"theme"`
	Selectivity    float64    `json:"selectivity"`
	ResultTime     string     `json:"result_time"`
	ResultMemory   string     `json:"result_memory"`
	ResultsCount   int        `json:"results_count"`
}

func SqlQueryToJSON(q repository.Request, creatorLogin string, moderatorLogin string, completedItemCount int) SqlQueryJSON {
	var mLogin *string
	if moderatorLogin != "" {
		mLogin = &moderatorLogin
	}
	return SqlQueryJSON{
		ID:             q.ID,
		Status:         q.Status,
		CreatedAt:      q.CreatedAt,
		CreatedByID:    q.CreatedByID,
		FormedAt:       q.FormedAt,
		CompletedAt:    q.FinishedAt,
		ModeratorID:    q.ModeratorID,
		CreatorLogin:   creatorLogin,
		ModeratorLogin: mLogin,
		QueryText:      q.QueryDescription,
		Theme:          q.QueryDescription,
		Selectivity:    q.Selectivity,
		ResultTime:     q.ResultTime,
		ResultMemory:   q.ResultMemory,
		ResultsCount:   completedItemCount,
	}
}

type SqlQueryItemJSON struct {
	IndexedTableID   string   `json:"indexed_table_id"`
	Name             string   `json:"name"`
	TableSize        string   `json:"table_size"`
	Selectivity      float64  `json:"selectivity"`
	ImageURL         string   `json:"image_url"`
	VideoURL         string   `json:"video_url"`
	Quantity         int      `json:"quantity"`
	Position         int      `json:"position"`
	CalculatedTimeMs *float64 `json:"calculated_time_ms"`
}

func SqlQueryItemToJSON(it repository.RequestService) SqlQueryItemJSON {
	var calc *float64
	if it.CalculatedTimeMs != 0 {
		v := it.CalculatedTimeMs
		calc = &v
	}
	videoKey := ""
	if it.Service != nil {
		videoKey = it.Service.GifKey
	}
	return SqlQueryItemJSON{
		IndexedTableID:   it.ServiceID,
		Name:             it.ServiceName,
		TableSize:        it.TableSize,
		Selectivity:      it.Selectivity,
		ImageURL:         mediaURL(it.ImageKey),
		VideoURL:         mediaURL(videoKey),
		Quantity:         it.Quantity,
		Position:         it.Position,
		CalculatedTimeMs: calc,
	}
}

func mediaURL(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	host := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if strings.TrimSpace(host) == "" {
		minioHost := os.Getenv("MINIO_HOST")
		if minioHost == "" {
			minioHost = "localhost"
		}
		minioPort := os.Getenv("MINIO_PORT")
		if minioPort == "" {
			minioPort = "9000"
		}
		host = fmt.Sprintf("http://%s:%s", minioHost, minioPort)
	}
	host = strings.TrimRight(host, "/")
	bucket := os.Getenv("MINIO_BUCKET")
	if strings.TrimSpace(bucket) == "" {
		bucket = "test"
	}
	return fmt.Sprintf("%s/%s/%s", host, bucket, key)
}

type CartJSON struct {
	ID    *uint `json:"id"`
	Count int   `json:"count"`
}
