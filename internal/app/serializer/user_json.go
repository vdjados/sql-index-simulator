package serializer

import "web_backend/internal/app/repository"

type UserJSON struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type RegisterUserJSON struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginJSON struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponseJSON struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

func UserToJSON(u repository.User) UserJSON {
	return UserJSON{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  u.Role,
	}
}
