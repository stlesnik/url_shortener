package services

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		longURL  string
		expected expected
	}{
		{
			name:     "good case",
			longURL:  "http://mbrgaoyhv.yandex",
			expected: expected{longURLStr: "http://mbrgaoyhv.yandex", error: ""},
		},
		{
			name:     "bad body",
			longURL:  ``,
			expected: expected{longURLStr: "", error: "didnt get url"},
		},
		{
			name:     "bad url",
			longURL:  "://mbrgaoyhv.yandex",
			expected: expected{longURLStr: "", error: "got incorrect url to shorten: url=://mbrgaoyhv.yandex, err=parse \"://mbrgaoyhv.yandex\": missing protocol scheme"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.longURL))
			req.Header.Add("Content-Type", "text/plain")

			longURLStr, err := GetLongURLFromReq(req)

			if tt.expected.error != "" {
				require.EqualError(t, err, tt.expected.error)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.longURLStr, longURLStr)
			}
		})
	}
}
