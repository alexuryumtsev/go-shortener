package handlers

import (
	"net/http"

	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/pg"
)

// PingHandler проверяет соединение с базой данных.
//
// Возвращает:
//   - В случае успеха:
//     Код: 200 OK
//   - В случае ошибки:
//     Код: 500 Internal Server Error - если соединение с БД не установлено
func PingHandler(repo storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, является ли хранилище экземпляром DatabaseStorage
		if dbRepo, ok := repo.(*pg.DatabaseStorage); ok {
			if err := dbRepo.Ping(r.Context()); err != nil {
				http.Error(w, "Database connection error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}
}
