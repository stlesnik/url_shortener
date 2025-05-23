package models

type GetURLDTO struct {
	OriginalURL string `db:"original_url"`
	IsDeleted   bool   `db:"is_deleted"`
}

// APIPrepareShortURL
type APIRequestPrepareShURL struct {
	LongURL string `json:"url"`
}

type APIResponsePrepareShURL struct {
	ShortURL string `json:"result"`
}

// APIPrepareBatchShortURL
type APIRequestPrepareBatchShURL struct {
	CorrelationID string `json:"correlation_id"`
	LongURL       string `json:"original_url"`
}

type APIResponsePrepareBatchShURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// APIGetUserURLs
type BaseURLDTO struct {
	ShortURLHash string `db:"short_url"`
	OriginalURL  string `db:"original_url"`
}
type BaseURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// APIDeleteUserURLs
type DeleteTask struct {
	URLHash string
	UserID  string
}
