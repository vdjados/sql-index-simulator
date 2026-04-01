package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type authUser struct {
	ID   uint
	Role string
}

func (u authUser) IsModerator() bool {
	return strings.EqualFold(u.Role, "moderator")
}

func (h *Handler) AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := extractToken(ctx)
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "description": "Требуется авторизация"})
			return
		}

		user, ttl, tokenErr := h.parseAndValidateToken(tokenString)
		if tokenErr != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "description": "Невалидный токен"})
			return
		}
		blacklisted, err := h.Repository.IsTokenBlacklisted(context.Background(), tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "description": "Ошибка проверки токена"})
			return
		}
		if blacklisted {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "description": "Токен отозван"})
			return
		}

		ctx.Set("auth_user", user)
		ctx.Set("auth_token", tokenString)
		ctx.Set("auth_token_ttl", ttl)
		ctx.Next()
	}
}

func (h *Handler) ModeratorOnly() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, ok := currentUser(ctx)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "description": "Требуется авторизация"})
			return
		}
		if !user.IsModerator() {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "description": "Недостаточно прав"})
			return
		}
		ctx.Next()
	}
}

func extractToken(ctx *gin.Context) string {
	authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	if cookieToken, err := ctx.Cookie("access_token"); err == nil && strings.TrimSpace(cookieToken) != "" {
		return strings.TrimSpace(cookieToken)
	}
	return ""
}

func (h *Handler) parseAndValidateToken(tokenString string) (authUser, time.Duration, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return authUser{}, 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return authUser{}, 0, jwt.ErrTokenInvalidClaims
	}

	var uid uint64
	switch v := claims["uid"].(type) {
	case float64:
		uid = uint64(v)
	case string:
		parsed, parseErr := strconv.ParseUint(v, 10, 64)
		if parseErr != nil {
			return authUser{}, 0, parseErr
		}
		uid = parsed
	default:
		return authUser{}, 0, jwt.ErrTokenInvalidClaims
	}
	role, _ := claims["role"].(string)
	if role == "" {
		role = "user"
	}
	ttl, err := getTokenTTL(claims)
	if err != nil {
		return authUser{}, 0, err
	}
	return authUser{ID: uint(uid), Role: role}, ttl, nil
}

func getTokenTTL(claims jwt.MapClaims) (time.Duration, error) {
	expValue, ok := claims["exp"]
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}
	var expUnix int64
	switch v := expValue.(type) {
	case float64:
		expUnix = int64(v)
	case int64:
		expUnix = v
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0, err
		}
		expUnix = parsed
	default:
		return 0, jwt.ErrTokenInvalidClaims
	}
	ttl := time.Until(time.Unix(expUnix, 0))
	if ttl < 0 {
		return 0, jwt.ErrTokenExpired
	}
	return ttl, nil
}

func currentUser(ctx *gin.Context) (authUser, bool) {
	val, ok := ctx.Get("auth_user")
	if !ok {
		return authUser{}, false
	}
	u, ok := val.(authUser)
	return u, ok
}
