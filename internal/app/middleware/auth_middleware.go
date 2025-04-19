package middleware

import (
	"net/http"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
)

const cookieName = "auth_token"

func AuthMiddleware(userService user.UserService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			// Кука отсутствует, создаем новый JWT
			token, err := userService.GenerateUserToken()
			if err != nil {
				http.Error(w, "Failed to generate token", http.StatusInternalServerError)
				return
			}
			setAuthCookie(w, token)
			r.AddCookie(&http.Cookie{Name: cookieName, Value: token})
		} else {
			// Проверяем JWT
			userID, err := userService.VerifyUserToken(cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			r.AddCookie(&http.Cookie{Name: cookieName, Value: userID})
		}
		next.ServeHTTP(w, r)
	})
}

// Устанавливаем JWT в куку
func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
}
