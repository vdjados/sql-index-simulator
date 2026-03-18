package serializer

import (
	"time"

	"web_backend/internal/app/repository"
)

type SqlQueryJSON struct {
	ID              uint       `json:"id"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	FormedAt        *time.Time `json:"formed_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	CreatorLogin    string     `json:"creator_login"`
	ModeratorLogin  *string    `json:"moderator_login"`
	QueryDescription string    `json:"query_description"`
	Selectivity     float64    `json:"selectivity"`
	ResultTime      string     `json:"result_time"`
	ResultMemory    string     `json:"result_memory"`
	CompletedItemCount int     `json:"completed_item_count"`
}

func SqlQueryToJSON(q repository.Request, creatorLogin string, moderatorLogin string, completedItemCount int) SqlQueryJSON {
	var mLogin *string
	if moderatorLogin != "" {
		mLogin = &moderatorLogin
	}
	return SqlQueryJSON{
		ID:                 q.ID,
		Status:             q.Status,
		CreatedAt:          q.CreatedAt,
		FormedAt:           q.FormedAt,
		FinishedAt:         q.FinishedAt,
		CreatorLogin:       creatorLogin,
		ModeratorLogin:     mLogin,
		QueryDescription:   q.QueryDescription,
		Selectivity:        q.Selectivity,
		ResultTime:         q.ResultTime,
		ResultMemory:       q.ResultMemory,
		CompletedItemCount: completedItemCount,
	}
}

type SqlQueryItemJSON struct {
	ServiceID        string  `json:"service_id"`
	ServiceName      string  `json:"service_name"`
	TableSize        string  `json:"table_size"`
	Selectivity      float64 `json:"selectivity"`
	ImageKey         string  `json:"image_key"`
	Quantity         int     `json:"quantity"`
	Position         int     `json:"position"`
	CalculatedTimeMs *float64 `json:"calculated_time_ms"`
}

func SqlQueryItemToJSON(it repository.RequestService) SqlQueryItemJSON {
	var calc *float64
	if it.CalculatedTimeMs != 0 {
		v := it.CalculatedTimeMs
		calc = &v
	}
	return SqlQueryItemJSON{
		ServiceID:        it.ServiceID,
		ServiceName:      it.ServiceName,
		TableSize:        it.TableSize,
		Selectivity:      it.Selectivity,
		ImageKey:         it.ImageKey,
		Quantity:         it.Quantity,
		Position:         it.Position,
		CalculatedTimeMs: calc,
	}
}

type CartJSON struct {
	ID    *uint `json:"id"`
	Count int   `json:"count"`
}

