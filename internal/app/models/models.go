package models

// URLModel представляет собой модель для хранения информации о URL.
// Содержит идентификатор, оригинальный URL, идентификатор пользователя и флаг удаления.
// Используется для работы с базой данных и хранения информации о сокращённых URL.
type URLModel struct {
	ID      string
	URL     string
	UserID  string
	Deleted bool
}

// URLBatchModel представляет собой модель для пакетной обработки URL.
// Используется при создании множества коротких URL за один запрос.
type URLBatchModel struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponseModel представляет собой модель ответа для пакетной обработки URL.
// Содержит идентификатор корреляции и сокращённый URL.
type BatchResponseModel struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// RequestBody определяет структуру входных данных.
type RequestBody struct {
	URL string `json:"url"`
}

// ResponseBody определяет структуру ответа.
type ResponseBody struct {
	ShortURL string `json:"result"`
}
