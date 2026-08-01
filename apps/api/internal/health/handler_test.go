package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_AllWithNilDependenciesReturnsServiceUnavailable(t *testing.T) {
	handler := Handler{}

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.All(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}
}
