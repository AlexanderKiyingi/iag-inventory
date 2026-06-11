package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"iag-inventory/backend/internal/config"
	"iag-inventory/backend/internal/middleware"
)

func init() { gin.SetMode(gin.TestMode) }

func testRouter() *gin.Engine {
	// Non-jwt PlatformAuth: AttachPrincipal passes through without setting a
	// principal, so RequireAuth on /api/v1 still fails closed (401).
	return NewRouter(RouterDeps{
		Cfg:          config.Config{ServiceName: "inventory", Environment: "test"},
		PlatformAuth: middleware.NewPlatformAuth(middleware.PlatformAuthOptions{Mode: "test"}),
	})
}

func TestHealth_OK(t *testing.T) {
	w := httptest.NewRecorder()
	testRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("/health = %d, want 200", w.Code)
	}
}

func TestOverview_RequiresAuth(t *testing.T) {
	w := httptest.NewRecorder()
	testRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("/api/v1/overview unauthenticated = %d, want 401", w.Code)
	}
}

func TestPlatformStatus_RequiresAuth(t *testing.T) {
	w := httptest.NewRecorder()
	testRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/platform/status", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("/api/v1/platform/status unauthenticated = %d, want 401", w.Code)
	}
}
