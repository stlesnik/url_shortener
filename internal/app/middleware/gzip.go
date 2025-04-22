package middleware

import (
	"compress/gzip"
	"github.com/stlesnik/url_shortener/cmd/logger"
	"net/http"
	"strings"
)

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

	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (gw *gzipResponseWriter) Write(b []byte) (int, error) {
	ct := gw.Header().Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html") {
		return gw.writer.Write(b)
	}
	return gw.ResponseWriter.Write(b)
}

func WithCompress(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Encoding") != "gzip" {
			next(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() {
			if err := gz.Close(); err != nil {
				logger.Sugaarz.Warnw("failed to close gzip writer", "err", err)
			}
		}()
		gw := &gzipResponseWriter{ResponseWriter: w, writer: gz}
		next(gw, r)
	}
}
