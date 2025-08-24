package models

// GetURLDTO represents a URL record.
type GetURLDTO struct {
	OriginalURL string `db:"original_url"`
	IsDeleted   bool   `db:"is_deleted"`
}

// APIPrepareShortURL
// APIRequestPrepareShURL represents a request to shorten a URL.
type APIRequestPrepareShURL struct {
	LongURL string `json:"url"`
}

// APIResponsePrepareShURL represents a response with a shortened URL.
type APIResponsePrepareShURL struct {
	ShortURL string `json:"result"`
}

// APIPrepareBatchShortURL
// APIRequestPrepareBatchShURL represents a batch shorten request.
type APIRequestPrepareBatchShURL struct {
	CorrelationID string `json:"correlation_id"`
	LongURL       string `json:"original_url"`
}

// APIResponsePrepareBatchShURL represents a batch shorten response.
type APIResponsePrepareBatchShURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// APIGetUserURLs
// BaseURLDTO represents a user's URL record.
type BaseURLDTO struct {
	ShortURLHash string `db:"short_url"`
	OriginalURL  string `db:"original_url"`
}

// BaseURLResponse represents a user's URL response.
type BaseURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// APIDeleteUserURLs
// DeleteTask represents a delete task for a user's URL.
type DeleteTask struct {
	URLHash string
	UserID  string
}

// fileConfig represents the JSON configuration file structure.
type FileConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	TrustedSubnet   string `json:"trusted_subnet"`
}
