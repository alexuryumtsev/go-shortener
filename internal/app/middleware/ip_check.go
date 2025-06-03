package middleware

import (
	"net"
	"net/http"
)

// IPCheckMiddleware проверяет, что IP адрес клиента находится в доверенной подсети
func IPCheckMiddleware(trustedSubnet string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Если доверенная подсеть не задана, блокируем доступ
		if trustedSubnet == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Парсим CIDR
		_, subnet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
			return
		}

		// Получаем IP адрес из заголовка X-Real-IP
		clientIP := r.Header.Get("X-Real-IP")
		if clientIP == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Парсим IP адрес клиента
		ip := net.ParseIP(clientIP)
		if ip == nil {
			http.Error(w, "Invalid client IP", http.StatusBadRequest)
			return
		}

		// Проверяем, что IP адрес находится в доверенной подсети
		if !subnet.Contains(ip) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}
