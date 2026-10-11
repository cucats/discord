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
	"golang.org/x/oauth2"
)

func TestBrokeredEntraToken(t *testing.T) {
	previousIssuer, previousAlias, previousClientID, previousSecret := config.KeycloakIssuer, config.KeycloakEntraAlias, config.KeycloakClientID, config.KeycloakClientSecret
	t.Cleanup(func() {
		config.KeycloakIssuer, config.KeycloakEntraAlias, config.KeycloakClientID, config.KeycloakClientSecret = previousIssuer, previousAlias, previousClientID, previousSecret
	})
	config.KeycloakEntraAlias = "microsoft"
	config.KeycloakClientID = "discord-test"
	config.KeycloakClientSecret = "test-secret"
	keycloakToken := &oauth2.Token{AccessToken: "keycloak-token"}

	for _, tc := range []struct {
		name      string
		v1Only    bool
		wantError bool
	}{
		{name: "V1", v1Only: true},
		{name: "V2"},
		{name: "denied", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/broker/microsoft/token" {
					http.NotFound(w, r)
					return
				}
				if tc.wantError {
					http.Error(w, "denied", http.StatusForbidden)
					return
				}
				if r.Method == http.MethodGet {
					if r.Header.Get("Authorization") != "Bearer keycloak-token" {
						t.Error("V1 request did not include the Keycloak token")
					}
				} else if r.Method == http.MethodPost {
					if tc.v1Only {
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					if err := r.ParseForm(); err != nil {
						t.Error(err)
					}
					if r.Form.Get("token") != "keycloak-token" || r.Form.Get("client_id") != "discord-test" || r.Form.Get("client_secret") != "test-secret" {
						t.Error("V2 request did not include the expected client credentials and Keycloak token")
					}
				} else {
					t.Errorf("unexpected method %q", r.Method)
				}
				json.NewEncoder(w).Encode(map[string]string{"access_token": "entra-token"})
			}))
			defer server.Close()
			config.KeycloakIssuer = server.URL

			got, err := role.GetBrokeredEntraToken(context.Background(), keycloakToken)
			if tc.wantError {
				if err == nil {
					t.Fatal("expected broker access to be denied")
				}
				return
			}
			if err != nil || got != "entra-token" {
				t.Fatalf("got token %q, error %v", got, err)
			}
		})
	}
}

type graphRewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (t graphRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	copy := req.Clone(req.Context())
	copy.URL.Scheme = t.target.Scheme
	copy.URL.Host = t.target.Host
	copy.Host = t.target.Host
	return t.base.RoundTrip(copy)
}

func TestGraphCollegeAcrossPages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer entra-token" {
			t.Error("Microsoft Graph request did not include the Entra token")
		}
		switch {
		case r.URL.Path == "/v1.0/me":
			json.NewEncoder(w).Encode(map[string]string{"userPrincipalName": "user@cam.ac.uk", "displayName": "Cambridge User"})
		case r.URL.Path == "/v1.0/me/memberOf" && r.URL.Query().Get("page") == "":
			json.NewEncoder(w).Encode(map[string]any{
				"value":           []map[string]string{{"id": "0cbcd7fb-1f17-48fc-ac3e-4a22131fa92d"}},
				"@odata.nextLink": "https://graph.microsoft.com/v1.0/me/memberOf?page=2",
			})
		case r.URL.Path == "/v1.0/me/memberOf" && r.URL.Query().Get("page") == "2":
			json.NewEncoder(w).Encode(map[string]any{"value": []map[string]string{{"id": "8ae6faff-1731-49e1-b786-6d8e08f9373a"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	previousTransport := http.DefaultTransport
	http.DefaultTransport = graphRewriteTransport{target: target, base: previousTransport}
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	user, err := role.GetGraphUserInfo(context.Background(), "entra-token")
	if err != nil {
		t.Fatal(err)
	}
	if user.Account != "user@cam.ac.uk" || !user.IsStudent || user.College != role.Trinity {
		t.Fatalf("unexpected Graph user information: %+v", user)
	}
}
