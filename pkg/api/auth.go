package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("a9f3Kd8xPq2Ws5Rv7Tn0Yz4Bm6Hj1Uc")

// getPasswordHash возвращает хэш пароля
func getPasswordHash(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

// signinHandler обрабатывает авторизацию пользователя
func signinHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	envPass := os.Getenv("TODO_PASSWORD")
	if envPass != req.Password {
		sendError(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	// JWT с хэшем пароля
	claims := jwt.MapClaims{
		"hash": getPasswordHash(envPass),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		sendError(w, "ошибка при создании токена", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"token": tokenString}, http.StatusOK)
}

// auth проверяет аутентификацию по JWT-токену из cookie
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			cookie, err := r.Cookie("token")
			if err != nil {
				sendError(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			jwtStr := cookie.Value
			token, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
				return secretKey, nil
			})
			if err != nil || !token.Valid {
				sendError(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				sendError(w, "Authentification required", http.StatusUnauthorized)
				return
			}

			hashFromToken, ok := claims["hash"].(string)
			if !ok || hashFromToken != getPasswordHash(pass) {
				sendError(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
