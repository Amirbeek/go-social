package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amirbeek/social/internal/auth"
	"github.com/amirbeek/social/internal/store"
	"github.com/amirbeek/social/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func NewTestAppApplication(t *testing.T) *application {
	t.Helper()

	//logger := zap.NewNop().Sugar()
	logger := zap.Must(zap.NewProduction()).Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewCacheMockStore()
	testAuth := &auth.TestAuthenticator{}

	return &application{
		logger:        logger,
		store:         mockStore,
		cacheStorage:  mockCacheStore,
		authenticator: testAuth,
	}
}

func executeRequest(req *http.Request, mux *chi.Mux) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
