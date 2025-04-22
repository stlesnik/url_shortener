package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithCompress(t *testing.T) {
	handler := WithCompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	resp := w.Result()
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	gr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, gr.Close())
	}()
	body, _ := io.ReadAll(gr)
	assert.JSONEq(t, `{"ok":true}`, string(body))
}

func TestWithDecompress(t *testing.T) {
	original := "http://example.com/some-url"
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(original))
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	handler := WithDecompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, r.Body.Close())
		}()

		assert.Equal(t, original, string(body))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
