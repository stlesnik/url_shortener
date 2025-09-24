package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/stlesnik/url_shortener/internal/app/middleware"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/app/repository"
	"github.com/stlesnik/url_shortener/internal/app/services"
	"github.com/stlesnik/url_shortener/internal/config"
)

// mockRepo is a minimal mock implementation of services.Storager and services.DBStorager for examples.
type mockRepo struct{}

func (m *mockRepo) Ping(_ context.Context) error { return nil }
func (m *mockRepo) SaveURL(_ context.Context, _, _, _ string) (bool, error) {
	return false, nil
}
func (m *mockRepo) GetURL(_ context.Context, _ string) (models.GetURLDTO, error) {
	return models.GetURLDTO{OriginalURL: "http://example.com"}, nil
}
func (m *mockRepo) Close() error { return nil }
func (m *mockRepo) GetURLList(_ context.Context, _ string) ([]models.BaseURLDTO, error) {
	return []models.BaseURLDTO{
		{ShortURLHash: "9uOVtk2tmuQ=", OriginalURL: "http://example.com"},
	}, nil
}
func (m *mockRepo) SaveBatchURL(_ context.Context, _ []repository.URLPair) error {
	return nil
}
func (m *mockRepo) DeleteURLList(values []interface{}, _ []string) (int64, error) {
	return int64(len(values)), nil
}
func (m *mockRepo) GetStats(_ context.Context) (models.StatsDTO, error) {
	return models.StatsDTO{URLCount: 1, UserCount: 1}, nil
}

// Example for SaveURL handler
func ExampleHandler_SaveURL() {
	repo := &mockRepo{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	h := &Handler{service: services.New(repo, cfg)}
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("http://example.com"))
	ctx := context.WithValue(req.Context(), middleware.UserIDKeyName, "test")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.SaveURL(w, req)
	resp := w.Result()
	err := resp.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println("status", resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("body: %s\n", string(body))
	// Output:
	// status 201
	// body: http://localhost:8080/9uOVtk2tmuQ=
}

// Example for GetLongURL handler
func ExampleHandler_GetLongURL() {
	repo := &mockRepo{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	h := &Handler{service: services.New(repo, cfg)}
	req := httptest.NewRequest(http.MethodGet, "/9uOVtk2tmuQ=", nil)
	w := httptest.NewRecorder()

	h.GetLongURL(w, req)
	resp := w.Result()
	err := resp.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println("status", resp.StatusCode)
	fmt.Printf("headers: Location: %s\n", resp.Header.Get("Location"))
	// Output:
	// status 307
	// headers: Location: http://example.com
}

// Example for APIPrepareShortURL handler
func ExampleHandler_APIPrepareShortURL() {
	repo := &mockRepo{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	h := &Handler{service: services.New(repo, cfg)}
	apiReq := models.APIRequestPrepareShURL{LongURL: "http://example.com"}
	body, _ := json.Marshal(apiReq)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.APIPrepareShortURL(w, req)
	resp := w.Result()
	err := resp.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println("status", resp.StatusCode)
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("body: %s\n", string(respBody))
	// Output:
	// status 201
	// body: {"result":"http://localhost:8080/9uOVtk2tmuQ="}
}

// Example for APIPrepareBatchShortURL handler
func ExampleHandler_APIPrepareBatchShortURL() {
	repo := &mockRepo{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	h := &Handler{service: services.New(repo, cfg)}
	batchReq := []models.APIRequestPrepareBatchShURL{{CorrelationID: "1", LongURL: "http://example.com"}}
	body, _ := json.Marshal(batchReq)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.APIPrepareBatchShortURL(w, req)
	resp := w.Result()
	err := resp.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println("status", resp.StatusCode)
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("body: %s\n", string(respBody))
	// Output:
	// status 201
	// body: [{"correlation_id":"1","short_url":"http://localhost:8080/9uOVtk2tmuQ="}]
}

// Example for APIGetUserURLs handler
func ExampleHandler_APIGetUserURLs() {
	repo := &mockRepo{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	h := &Handler{service: services.New(repo, cfg)}
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKeyName, "test")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.APIGetUserURLs(w, req)
	resp := w.Result()
	err := resp.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println("status", resp.StatusCode)
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("body: %s\n", string(respBody))
	// Output:
	// status 200
	// body: [{"short_url":"http://localhost:8080/9uOVtk2tmuQ=","original_url":"http://example.com"}]
}

// Example for APIDeleteUserURLs handler
func ExampleHandler_APIDeleteUserURLs() {
	repo := &mockRepo{}
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	h := &Handler{service: services.New(repo, cfg)}
	urlHashes := []string{"9uOVtk2tmuQ="}
	body, _ := json.Marshal(urlHashes)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKeyName, "test")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.APIDeleteUserURLs(w, req)
	resp := w.Result()
	err := resp.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println("status", resp.StatusCode)
	// Output:
	// status 202
}
