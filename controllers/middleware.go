package controllers

import (
	"clothes/models"
	"compress/gzip"
	"context"
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
		c, err := r.Cookie("session_token")
		if err != nil || c.Value == "" {
			slog.Info("No session cookie found, redirecting to login")
			http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
			return
		}

		siteUser, err := models.ApiQuery[models.SiteUser](r.Context(), "user_validate_session", c.Value)
		if err != nil {
			slog.Error("Error validating session token", "error", err)
			// delete cookie
			http.SetCookie(w, &http.Cookie{
				Name:   "session_token",
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
			http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "siteUser", *siteUser))

		next.ServeHTTP(w, r)
	})
}
