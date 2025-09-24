package middleware

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stlesnik/url_shortener/internal/config"
)

func TestWithTrustedSubnet(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		expectedStatus int
		expectedBody   string
		nextCalled     bool
	}{
		{
			name:           "empty trusted subnet",
			trustedSubnet:  "",
			xRealIP:        "192.168.1.1",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "IP address is not in trusted Subnet",
			nextCalled:     false,
		},
		{
			name:           "invalid subnet",
			trustedSubnet:  "invalid_subnet",
			xRealIP:        "192.168.1.1",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Cannot parse subnet",
			nextCalled:     false,
		},
		{
			name:           "missing X-Real-IP header",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Cannot find IP address",
			nextCalled:     false,
		},
		{
			name:           "invalid IP address",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "invalid_ip",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Cannot find IP address",
			nextCalled:     false,
		},
		{
			name:           "IP in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.10",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
			nextCalled:     true,
		},
		{
			name:           "IP not in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "IP address is not in trusted Subnet",
			nextCalled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				TrustedSubnet: tt.trustedSubnet,
			}

			req := httptest.NewRequest("GET", "/", nil)
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			rr := httptest.NewRecorder()

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("OK"))
				if err != nil {
					require.NoError(t, err)
				}
			})

			middleware := WithTrustedSubnet(cfg, nextHandler)
			middleware.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			body := strings.TrimSpace(rr.Body.String())
			if body != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got '%v' want '%v'",
					body, tt.expectedBody)
			}

			if nextCalled != tt.nextCalled {
				t.Errorf("next handler called: got %v want %v",
					nextCalled, tt.nextCalled)
			}
		})
	}
}
