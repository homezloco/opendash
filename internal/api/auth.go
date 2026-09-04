package api

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/opendash-project/opendash/internal/auth"
)

const (
	sessionCookie  = "opendash_session"
	loginLimit     = 5
	loginWindow    = time.Minute
	maxLimiterKeys = 10_000
)

type loginAttempt struct {
	count int
	reset time.Time
}

type loginLimiter struct {
	mu       sync.Mutex
	attempts map[string]loginAttempt
	now      func() time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{attempts: make(map[string]loginAttempt), now: time.Now}
}

func (l *loginLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if len(l.attempts) >= maxLimiterKeys {
		for candidate, attempt := range l.attempts {
			if !now.Before(attempt.reset) {
				delete(l.attempts, candidate)
			}
		}
		// Fail closed for new clients while all bounded slots are active.
		if _, exists := l.attempts[key]; !exists && len(l.attempts) >= maxLimiterKeys {
			return false
		}
	}
	attempt := l.attempts[key]
	if !now.Before(attempt.reset) {
		attempt = loginAttempt{reset: now.Add(loginWindow)}
	}
	if attempt.count >= loginLimit {
		return false
	}
	attempt.count++
	l.attempts[key] = attempt
	return true
}

func (l *loginLimiter) reset(key string) {
	l.mu.Lock()
	delete(l.attempts, key)
	l.mu.Unlock()
}

func (h *Handlers) ConfigureAuth(service *auth.Service, enabled, secure bool, lifetime time.Duration) {
	h.auth = service
	h.authEnabled = enabled
	h.secureCookies = secure
	h.sessionLifetime = lifetime
	h.limiter = newLoginLimiter()
}

func (h *Handlers) BootstrapStatus(w http.ResponseWriter, r *http.Request) {
	if !h.authEnabled {
		RespondJSON(w, http.StatusOK, map[string]bool{"authEnabled": false, "bootstrapRequired": false})
		return
	}
	required, err := h.auth.BootstrapRequired(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "store_error", "unable to read bootstrap status")
		return
	}
	RespondJSON(w, http.StatusOK, map[string]bool{"authEnabled": true, "bootstrapRequired": required})
}

func (h *Handlers) Bootstrap(w http.ResponseWriter, r *http.Request) {
	if !h.authAvailable(w) {
		return
	}
	client := clientIP(r)
	if !h.limiter.allow(client) {
		RespondError(w, http.StatusTooManyRequests, "rate_limited", "too many authentication attempts")
		return
	}
	var in credentials
	if err := decodeJSON(w, r, &in); err != nil {
		return
	}
	token, session, err := h.auth.Bootstrap(r.Context(), in.Username, in.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidInput):
			RespondError(w, http.StatusBadRequest, "invalid_credentials", err.Error())
		case errors.Is(err, auth.ErrAdminExists):
			RespondError(w, http.StatusConflict, "already_bootstrapped", "administrator already exists")
		default:
			RespondError(w, http.StatusInternalServerError, "auth_error", "unable to create administrator")
		}
		return
	}
	h.limiter.reset(client)
	h.setCookie(w, token, session.ExpiresAt)
	RespondJSON(w, http.StatusCreated, session)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if !h.authAvailable(w) {
		return
	}
	client := clientIP(r)
	if !h.limiter.allow(client) {
		RespondError(w, http.StatusTooManyRequests, "rate_limited", "too many login attempts")
		return
	}
	var in credentials
	if err := decodeJSON(w, r, &in); err != nil {
		return
	}
	token, session, err := h.auth.Login(r.Context(), in.Username, in.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			RespondError(w, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
		} else {
			RespondError(w, http.StatusInternalServerError, "auth_error", "unable to start session")
		}
		return
	}
	h.limiter.reset(client)
	h.setCookie(w, token, session.ExpiresAt)
	RespondJSON(w, http.StatusOK, session)
}

func (h *Handlers) Session(w http.ResponseWriter, r *http.Request) {
	if !h.authAvailable(w) {
		return
	}
	session, err := h.currentSession(r)
	if err != nil {
		h.clearCookie(w)
		RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	RespondJSON(w, http.StatusOK, session)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if !h.authAvailable(w) {
		return
	}
	cookie, _ := r.Cookie(sessionCookie)
	if cookie != nil {
		if err := h.auth.Logout(r.Context(), cookie.Value); err != nil {
			RespondError(w, http.StatusInternalServerError, "auth_error", "unable to end session")
			return
		}
	}
	h.clearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) authAvailable(w http.ResponseWriter) bool {
	if h.authEnabled {
		return true
	}
	RespondError(w, http.StatusNotFound, "not_found", "resource not found")
	return false
}

func (h *Handlers) setCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		Expires:  expires.UTC(),
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *Handlers) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Path:     "/",
		Expires:  time.Unix(1, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *Handlers) currentSession(r *http.Request) (auth.Session, error) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return auth.Session{}, err
	}
	return h.auth.Authenticate(r.Context(), cookie.Value)
}

func (h *Handlers) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.authEnabled {
			next.ServeHTTP(w, r)
			return
		}
		session, err := h.currentSession(r)
		if err != nil {
			h.clearCookie(w)
			RespondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}
		if requiresCSRF(r.Method) && !equalToken(r.Header.Get("X-CSRF-Token"), session.CSRFToken) {
			RespondError(w, http.StatusForbidden, "csrf_invalid", "valid CSRF token required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requiresCSRF(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}

func equalToken(provided, expected string) bool {
	return len(provided) == len(expected) && subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

// clientIP deliberately trusts only RemoteAddr. Forwarded headers are attacker-controlled
// unless the application has an explicit trusted-proxy configuration.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	mediaType := strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0])
	if !strings.EqualFold(mediaType, "application/json") {
		RespondError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "application/json required")
		return errors.New("content type")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		RespondError(w, http.StatusBadRequest, "invalid_json", "request body must contain one JSON object")
		return errors.New("trailing data")
	}
	return nil
}
