package models

type APIRequestPrepareShURL struct {
	LongURL string `json:"url"`
}

type APIResponsePrepareShURL struct {
	ShortURL string `json:"result"`
}
