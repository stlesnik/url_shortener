package models

type ApiRequestPrepareShURL struct {
	LongURL string `json:"url"`
}

type ApiResponsePrepareShURL struct {
	ShortURL string `json:"result"`
}
