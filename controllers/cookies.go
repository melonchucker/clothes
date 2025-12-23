package controllers

import (
	"clothes/models"
	"clothes/views/widgets"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
)

const (
	sessionCookieName = "session_token"
	alertCookieName   = "alert_message"
)

type alert struct {
	Level   widgets.AlertLevel `json:"level"`
	Message string             `json:"message"`
}

func DecodeJSONCookie[T any](r *http.Request, name string) (*T, error) {
	c, err := r.Cookie(name)
	if err != nil {
		return nil, err
	}

	raw, err := base64.RawURLEncoding.DecodeString(c.Value)
	if err != nil {
		return nil, errors.New("cookie is not valid base64")
	}

	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, errors.New("cookie is not valid json")
	}

	return &out, nil
}

func EncodeJSONCookie[T any](w http.ResponseWriter, name string, value *T) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	encoded := base64.RawURLEncoding.EncodeToString(raw)
	http.SetCookie(w, &http.Cookie{
		Name:  name,
		Value: encoded,
	})
	return nil
}

func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:   name,
		Value:  "",
		MaxAge: -1,
	})
}

func setAlert(w http.ResponseWriter, alertType widgets.AlertLevel, alertMessage string) error {
	return EncodeJSONCookie(w, alertCookieName, &alert{
		Level:   alertType,
		Message: alertMessage,
	})
}

func getAndClearAlert(r *http.Request, w http.ResponseWriter) (*alert, error) {
	a, err := DecodeJSONCookie[alert](r, alertCookieName)
	if err != nil {
		return nil, err
	}

	if a == nil {
		return nil, nil
	}

	ClearCookie(w, alertCookieName)

	return &alert{
		Level:   a.Level,
		Message: a.Message,
	}, nil
}

func setSession(r *http.Request, w http.ResponseWriter, email string, password string) error {
	session, err := models.ApiQuery[string](r.Context(), "site_user_authenticate", email, password)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    *session,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func clearSession(w http.ResponseWriter, r *http.Request) error {
	c, err := r.Cookie(sessionCookieName)
	if err == http.ErrNoCookie {
		return nil
	} else if err != nil {
		return err
	}

	_, err = models.ApiQuery[string](r.Context(), "user_signout", c.Value)
	if err != nil {
		panic(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func getSession(w http.ResponseWriter, r *http.Request) (*models.SiteUser, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, err
	}

	siteUser, err := models.ApiQuery[models.SiteUser](r.Context(), "user_validate_session", c.Value)
	if err != nil {
		clearSession(w, r)
		return nil, err
	}

	return siteUser, nil
}
