package models

// GetURLDTO represents the data transfer object for a URL record.
type GetURLDTO struct {
	OriginalURL string `db:"original_url"`
	IsDeleted   bool   `db:"is_deleted"`
}

// APIPrepareShortURL
// APIRequestPrepareShURL represents the request body for the API endpoint to shorten a URL.
type APIRequestPrepareShURL struct {
	LongURL string `json:"url"`
}

// APIResponsePrepareShURL represents the response body for the API endpoint to shorten a URL.
type APIResponsePrepareShURL struct {
	ShortURL string `json:"result"`
}

// APIPrepareBatchShortURL
// APIRequestPrepareBatchShURL represents the request body for the batch URL shortening API endpoint.
type APIRequestPrepareBatchShURL struct {
	CorrelationID string `json:"correlation_id"`
	LongURL       string `json:"original_url"`
}

// APIResponsePrepareBatchShURL represents the response body for the batch URL shortening API endpoint.
type APIResponsePrepareBatchShURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// APIGetUserURLs
// BaseURLDTO represents a base URL record for a user.
type BaseURLDTO struct {
	ShortURLHash string `db:"short_url"`
	OriginalURL  string `db:"original_url"`
}

// BaseURLResponse represents the response object for a user's URL.
type BaseURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// APIDeleteUserURLs
// DeleteTask represents a task to delete a user's URL.
type DeleteTask struct {
	URLHash string
	UserID  string
}
