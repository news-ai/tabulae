package tabulae

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleIndexReturnsWithStatusOK(t *testing.T) {
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Response body did not contain expected %v:\n\tbody: %v", "200", response.Code)
	}
}

func TestHandleIndexContainsAmsterdam(t *testing.T) {
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(response, request)

	body := response.Body.String()
	if !strings.Contains(body, "Amsterdam") {
		t.Fatalf("Response body did not contain expected %v:\n\tbody: %v", "Amsterdam", body)
	}
}
