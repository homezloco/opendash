package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/opendash-project/opendash/internal/auth"
	"github.com/opendash-project/opendash/internal/runtime/noop"
	"github.com/opendash-project/opendash/internal/store"
)

func TestAuthenticationFlow(t *testing.T) {
	s, err := store.OpenSQLite(context.Background(), t.TempDir()+"/opendash.db", true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h := NewHandlers(s, noop.New())
	h.ConfigureAuth(auth.New(s.DB(), time.Hour), true, false, time.Hour)
	server := httptest.NewServer(NewRouter(h, 1<<20))
	defer server.Close()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	client.Jar = jar

	assertStatus(t, client, http.MethodGet, server.URL+"/api/v1/summary", "", "", http.StatusUnauthorized)
	response := assertStatus(t, client, http.MethodPost, server.URL+"/api/v1/bootstrap", `{"username":"admin","password":"correct horse battery staple"}`, "", http.StatusCreated)
	cookies := response.Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode || cookies[0].Expires.IsZero() {
		t.Fatal("secure cookie attributes missing")
	}
	var session auth.Session
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	assertStatus(t, client, http.MethodPost, server.URL+"/api/v1/auth/logout", "", "", http.StatusForbidden)
	assertStatus(t, client, http.MethodPost, server.URL+"/api/v1/auth/logout", "", session.CSRFToken, http.StatusNoContent)
	assertStatus(t, client, http.MethodGet, server.URL+"/api/v1/auth/session", "", "", http.StatusUnauthorized)

	response = assertStatus(t, client, http.MethodPost, server.URL+"/api/v1/auth/login", `{"username":"admin","password":"correct horse battery staple"}`, "", http.StatusOK)
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	assertStatus(t, client, http.MethodGet, server.URL+"/api/v1/auth/session", "", "", http.StatusOK)
	assertStatus(t, client, http.MethodPost, server.URL+"/api/v1/auth/logout", "", session.CSRFToken, http.StatusNoContent)
}

func TestAuthenticationDisabled(t *testing.T) {
	s, err := store.OpenSQLite(context.Background(), t.TempDir()+"/opendash.db", true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h := NewHandlers(s, noop.New())
	h.ConfigureAuth(auth.New(s.DB(), time.Hour), false, false, time.Hour)
	server := httptest.NewServer(NewRouter(h, 1<<20))
	defer server.Close()
	client := server.Client()

	response := assertStatus(t, client, http.MethodGet, server.URL+"/api/v1/bootstrap", "", "", http.StatusOK)
	var status map[string]bool
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if status["authEnabled"] || status["bootstrapRequired"] {
		t.Fatalf("unexpected disabled status: %#v", status)
	}
	assertStatus(t, client, http.MethodGet, server.URL+"/api/v1/summary", "", "", http.StatusOK)
	assertStatus(t, client, http.MethodPost, server.URL+"/api/v1/auth/login", `{}`, "", http.StatusNotFound)
}

func assertStatus(t *testing.T, client *http.Client, method, url, body, csrf string, want int) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if csrf != "" {
		request.Header.Set("X-CSRF-Token", csrf)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != want {
		response.Body.Close()
		t.Fatalf("%s %s: got %d, want %d", method, url, response.StatusCode, want)
	}
	return response
}
