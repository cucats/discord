package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cucats.org/discord/role"
)

func TestKeycloakCallbackWithoutSessionFails(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/keycloak/callback", nil)
	role.New(nil).KeycloakCallback(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("got HTTP %d, want 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Verification Failed") || !strings.Contains(w.Body.String(), "Session expired") || strings.Contains(w.Body.String(), "Verification Successful") {
		t.Fatal("callback error rendered as success")
	}
}
