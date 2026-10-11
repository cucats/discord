package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"cucats.org/discord/config"
	"cucats.org/discord/role"
)

func TestStoreVerificationInKeycloak(t *testing.T) {
	previousIssuer, previousID, previousSecret := config.KeycloakIssuer, config.KeycloakWriterClientID, config.KeycloakWriterClientSecret
	t.Cleanup(func() {
		config.KeycloakIssuer, config.KeycloakWriterClientID, config.KeycloakWriterClientSecret = previousIssuer, previousID, previousSecret
	})

	user := map[string]any{
		"id": "keycloak-user", "username": "cambridge-user",
		"attributes": map[string][]string{"existing": {"keep me"}, "college": {"Old college"}},
	}
	var writes int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/realms/master/protocol/openid-connect/token":
			if r.Method != http.MethodPost {
				t.Errorf("token method: %s", r.Method)
			}
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			if r.Form != nil && (r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("client_id") != "writer" || r.Form.Get("client_secret") != "secret") {
				t.Errorf("unexpected token request: %s", url.Values{"grant_type": {r.Form.Get("grant_type")}, "client_id": {r.Form.Get("client_id")}}.Encode())
			}
			json.NewEncoder(w).Encode(map[string]string{"access_token": "writer-token"})
		case "/admin/realms/master/users/keycloak-user":
			if r.Header.Get("Authorization") != "Bearer writer-token" {
				t.Errorf("missing writer authentication")
			}
			switch r.Method {
			case http.MethodGet:
				json.NewEncoder(w).Encode(user)
			case http.MethodPut:
				writes++
				if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
					t.Error(err)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				http.NotFound(w, r)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	config.KeycloakIssuer = server.URL + "/realms/master"
	config.KeycloakWriterClientID = "writer"
	config.KeycloakWriterClientSecret = "secret"

	if err := role.StoreVerificationInKeycloak(context.Background(), "keycloak-user", "discord-123", "first-name", role.Trinity); err != nil {
		t.Fatal(err)
	}
	attributes := user["attributes"].(map[string]any)
	if attributes["existing"].([]any)[0] != "keep me" ||
		attributes["discord_id"].([]any)[0] != "discord-123" ||
		attributes["discord_username"].([]any)[0] != "first-name" ||
		attributes["college"].([]any)[0] != "Trinity" {
		t.Fatalf("wrong attributes: %+v", attributes)
	}
	if err := role.StoreVerificationInKeycloak(context.Background(), "keycloak-user", "discord-123", "renamed", role.Unknown); err != nil {
		t.Fatal(err)
	}
	attributes = user["attributes"].(map[string]any)
	if _, exists := attributes["college"]; exists || attributes["discord_username"].([]any)[0] != "renamed" || writes != 2 {
		t.Fatalf("stale verification: %+v, writes=%d", attributes, writes)
	}
}
