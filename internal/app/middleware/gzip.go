package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/stlesnik/url_shortener/internal/logger"
)

// WithDecompress is a middleware that decompresses gzip-encoded requests.
func WithDecompress(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			next(w, r)
			return
		}
		gr, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "failed to decompress request", http.StatusBadRequest)
			return
		}
		defer func() {
			if err := gr.Close(); err != nil {
				logger.Sugaarz.Warnw("failed to close gzip reader", "err", err)
			}
		}()
		r.Body = gr
		r.Header.Del("Content-Encoding")
		next(w, r)
	}
}

// gzipResponseWriter wraps http.ResponseWriter for gzip.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	wroteHeader bool
}

// WriteHeader writes the HTTP header and sets up gzip if needed.
func (gw *gzipResponseWriter) WriteHeader(status int) {
	if !gw.wroteHeader {
		ct := gw.Header().Get("Content-Type")
		if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/") {
			gw.Header().Add("Content-Encoding", "gzip")
			gw.writer = gzip.NewWriter(gw.ResponseWriter)
		}
		gw.wroteHeader = true
	}
	gw.ResponseWriter.WriteHeader(status)
}

// Write writes the response body, compressing it if gzip is enabled.
func (gw *gzipResponseWriter) Write(b []byte) (int, error) {
	if !gw.wroteHeader {
		gw.WriteHeader(http.StatusOK)
	}

	if gw.writer != nil {
		return gw.writer.Write(b)
	}
	return gw.ResponseWriter.Write(b)
}

// WithCompress is a middleware that compresses responses using gzip if supported by the client.
func WithCompress(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}
		gw := &gzipResponseWriter{ResponseWriter: w}
		next(gw, r)

		if gw.writer != nil {
			_ = gw.writer.Close()
		}
	}
}
