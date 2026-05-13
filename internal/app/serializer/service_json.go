package serializer

import (
	"strings"
	"web_backend/internal/app/repository"
)

type ServiceJSON struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	TableSize          string `json:"table_size"`
	Speed              string `json:"speed"`
	Description        string `json:"description"`
	ShortDescriptionEN string `json:"short_description_en"`
	ImageKey           string `json:"image_key"`
	VideoKey           string `json:"video_key"`
	ImageURL           string `json:"image_url"`
	VideoURL           string `json:"video_url"`
}

func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || s == "" {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return strings.TrimSpace(string(r[:max]))
}

// effectiveShortDescriptionEN never returns empty for a valid row: DB column may be blank for legacy rows.
func effectiveShortDescriptionEN(s repository.Service) string {
	if v := strings.TrimSpace(s.ShortDescriptionEN); v != "" {
		return truncateRunes(v, 500)
	}
	if v := strings.TrimSpace(s.Description); v != "" {
		return truncateRunes(v, 120)
	}
	return truncateRunes(strings.TrimSpace(s.Name), 120)
}

func ServiceToJSON(s repository.Service) ServiceJSON {
	return ServiceJSON{
		ID:                 s.ID,
		Name:               s.Name,
		TableSize:          s.TableSize,
		Speed:              s.Speed,
		Description:        s.Description,
		ShortDescriptionEN: effectiveShortDescriptionEN(s),
		ImageKey:           s.ImageKey,
		VideoKey:           s.GifKey, // используем GifKey как key короткого видео (по заданию)
		ImageURL:           mediaURL(s.ImageKey),
		VideoURL:           mediaURL(s.GifKey),
	}
}
