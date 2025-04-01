package services

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateShortKey(t *testing.T) {
	tests := []struct {
		name    string
		longURL string
		want    string
	}{
		{
			"generate hash",
			"http://mbrgaoyhv.yandex",
			"_SGMGLQIsIM=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GenerateShortKey(tt.longURL))
		})
	}
}

func TestProcessRequest(t *testing.T) {
	type expected struct {
		id     string
		method string
		error  string
		status int
	}
	tests := []struct {
		name     string
		method   string
		path     string
		expected expected
	}{
		{
			name:     "valid GET request",
			method:   http.MethodGet,
			path:     "/abc123",
			expected: expected{id: "abc123", method: http.MethodGet, status: http.StatusOK},
		},
		{
			name:   "invalid path with slash",
			method: http.MethodGet,
			path:   "/invalid/id/",
			expected: expected{
				method: "",
				error:  "incorrect url",
				status: http.StatusBadRequest,
			},
		},
		{
			name:   "invalid method PUT",
			method: http.MethodPut,
			path:   "/valid",
			expected: expected{
				method: "",
				error:  "incorrect method: only GET and POST allowed",
				status: http.StatusBadRequest,
			},
		},
		{
			name:   "valid POST request",
			method: http.MethodPost,
			path:   "/",
			expected: expected{
				method: http.MethodPost,
				status: http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			id, method, err := ProcessRequest(rec, req)

			assert.Equal(t, tt.expected.id, id)
			assert.Equal(t, tt.expected.method, method)

			assert.Equal(t, tt.expected.status, rec.Code)

			if tt.expected.error != "" {
				assert.ErrorContains(t, err, tt.expected.error)
				assert.Contains(t, rec.Body.String(), tt.expected.error+"\n")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
