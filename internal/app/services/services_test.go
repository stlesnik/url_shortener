package services

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestGetLongURL(t *testing.T) {
	type expected struct {
		longURLStr string
		error      string
	}
	tests := []struct {
		name     string
		body     io.Reader
		expected expected
	}{
		{
			name:     "good case",
			body:     strings.NewReader(`url=http%3A%2F%2Fmbrgaoyhv.yandex`),
			expected: expected{longURLStr: "http://mbrgaoyhv.yandex", error: ""},
		},
		{
			name:     "bad body",
			body:     strings.NewReader(``),
			expected: expected{longURLStr: "", error: "didnt get url"},
		},
		{
			name:     "bad url",
			body:     strings.NewReader(`url=%3A%2F%2Fmbrgaoyhv.yandex%2F123%2F4%2F5"}`),
			expected: expected{longURLStr: "", error: "got incorrect url to shorten: url=://mbrgaoyhv.yandex/123/4/5\"}, err=parse \"://mbrgaoyhv.yandex/123/4/5\\\"}\": missing protocol scheme"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", tt.body)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			longURLStr, err := GetLongURL(req)

			if tt.expected.error != "" {
				require.EqualError(t, err, tt.expected.error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.longURLStr, longURLStr)
			}
		})
	}
}
