package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aniliberty-indexer/internal/config"
)

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	HandleRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected application/json, got %s", contentType)
	}
}

func TestHandleCaps(t *testing.T) {
	req := httptest.NewRequest("GET", "/api?t=caps", nil)
	w := httptest.NewRecorder()

	HandleRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<caps>") {
		t.Errorf("expected caps XML in body, got: %s", body)
	}
}

func TestHandleUnknownTorznabAction(t *testing.T) {
	req := httptest.NewRequest("GET", "/api?t=invalidaction", nil)
	w := httptest.NewRecorder()

	HandleRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, `error code="201"`) {
		t.Errorf("expected error code 201, got: %s", body)
	}
}

func TestCheckAPIKey(t *testing.T) {
	origKey := config.API_KEY
	defer func() { config.API_KEY = origKey }()

	config.API_KEY = "secret123"

	// Request without API key
	reqNoKey := httptest.NewRequest("GET", "/api?t=caps", nil)
	wNoKey := httptest.NewRecorder()
	HandleRequest(wNoKey, reqNoKey)
	if wNoKey.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 when API key is missing, got %d", wNoKey.Result().StatusCode)
	}

	// Request with valid API key in query
	reqQueryKey := httptest.NewRequest("GET", "/api?t=caps&apikey=secret123", nil)
	wQueryKey := httptest.NewRecorder()
	HandleRequest(wQueryKey, reqQueryKey)
	if wQueryKey.Result().StatusCode != http.StatusOK {
		t.Errorf("expected 200 with valid apikey query param, got %d", wQueryKey.Result().StatusCode)
	}

	// Request with valid API key in header
	reqHeaderKey := httptest.NewRequest("GET", "/api?t=caps", nil)
	reqHeaderKey.Header.Set("X-Api-Key", "secret123")
	wHeaderKey := httptest.NewRecorder()
	HandleRequest(wHeaderKey, reqHeaderKey)
	if wHeaderKey.Result().StatusCode != http.StatusOK {
		t.Errorf("expected 200 with valid X-Api-Key header, got %d", wHeaderKey.Result().StatusCode)
	}
}

func TestParseCategories(t *testing.T) {
	req := httptest.NewRequest("GET", "/api?cat=5000,5070&cat=2040", nil)
	cats := parseCategories(req.URL.Query())
	expected := []int{5000, 5070, 2040}

	if len(cats) != len(expected) {
		t.Fatalf("expected %d categories, got %d", len(expected), len(cats))
	}
	for i, c := range cats {
		if c != expected[i] {
			t.Errorf("expected cat[%d]=%d, got %d", i, expected[i], c)
		}
	}
}

