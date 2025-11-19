package controllers

import (
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request", "method", r.Method, "path", r.URL, "remote", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.Header().Set("Content-Encoding", "gzip")
	w.ResponseWriter.WriteHeader(statusCode)
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := &gzipResponseWriter{ResponseWriter: w, writer: gz}
		next.ServeHTTP(gzw, r)
	})
}

func authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sessionInfo struct {
			Email        string `json:"email"`
			SessionToken string `json:"session_token"`
		}
		var decoded []byte

		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			// redirect to login
			slog.Info("No session cookie found, redirecting to login")
			http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
			return
		}

		decoded, err = base64.StdEncoding.DecodeString(cookie.Value)
		if err != nil {
			return
		}

		json.Unmarshal(decoded, &sessionInfo)

		if err := json.Unmarshal(decoded, &sessionInfo); err != nil {
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "email", sessionInfo.Email))

		next.ServeHTTP(w, r)
	})
}
