package serializer

import "web_backend/internal/app/repository"

type ServiceJSON struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	TableSize   string `json:"table_size"`
	Speed       string `json:"speed"`
	Description string `json:"description"`
	ShortDescriptionEN string `json:"short_description_en"`
	ImageKey    string `json:"image_key"`
	VideoKey    string `json:"video_key"`
	ImageURL    string `json:"image_url"`
	VideoURL    string `json:"video_url"`
}

func ServiceToJSON(s repository.Service) ServiceJSON {
	return ServiceJSON{
		ID:          s.ID,
		Name:        s.Name,
		TableSize:   s.TableSize,
		Speed:       s.Speed,
		Description: s.Description,
		ShortDescriptionEN: s.ShortDescriptionEN,
		ImageKey:    s.ImageKey,
		VideoKey:    s.GifKey, // используем GifKey как key короткого видео (по заданию)
		ImageURL:    mediaURL(s.ImageKey),
		VideoURL:    mediaURL(s.GifKey),
	}
}
