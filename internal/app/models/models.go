package models

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
