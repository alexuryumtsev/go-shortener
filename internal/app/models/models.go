package models

// URLMapping структура для хранения URL и его сокращённого идентификатора.
type URLModel struct {
	ID      string
	URL     string
	UserID  string
	Deleted bool
}
type URLBatchModel struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

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
