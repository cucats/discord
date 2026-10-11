package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cucats.org/discord/config"
	"cucats.org/discord/role"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"golang.org/x/oauth2"
)

func TestGetKeycloakUserInfo(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	key := jose.JSONWebKey{Key: privateKey, KeyID: "test-key", Algorithm: string(jose.RS256), Use: "sig"}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var userinfoSubject atomic.Value
	userinfoSubject.Store("cambridge-user")
	var issuer string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]string{
				"issuer": issuer, "authorization_endpoint": issuer + "/auth",
				"token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys",
				"userinfo_endpoint": issuer + "/userinfo",
			})
		case "/keys":
			json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{key.Public()}})
		case "/userinfo":
			if r.Header.Get("Authorization") != "Bearer access-token" {
				t.Errorf("missing bearer token: %q", r.Header.Get("Authorization"))
			}
			json.NewEncoder(w).Encode(map[string]any{
				"sub": userinfoSubject.Load().(string), "email": "user@cam.ac.uk",
				"name": "Cambridge User",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	issuer = server.URL

	previousIssuer, previousClientID, previousClientSecret, previousHost := config.KeycloakIssuer, config.KeycloakClientID, config.KeycloakClientSecret, config.Host
	t.Cleanup(func() {
		config.KeycloakIssuer, config.KeycloakClientID, config.KeycloakClientSecret, config.Host = previousIssuer, previousClientID, previousClientSecret, previousHost
	})
	config.KeycloakIssuer = issuer
	config.KeycloakClientID = "discord-test"
	config.KeycloakClientSecret = "test-secret"
	config.Host = issuer
	if err := role.InitKeycloakOAuth(context.Background()); err != nil {
		t.Fatal(err)
	}

	idToken := func(nonce string) string {
		t.Helper()
		raw, err := jwt.Signed(signer).Claims(map[string]any{
			"iss": issuer, "aud": "discord-test", "sub": "cambridge-user",
			"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "nonce": nonce,
		}).Serialize()
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	token := func(raw string) *oauth2.Token {
		return (&oauth2.Token{AccessToken: "access-token", TokenType: "Bearer"}).WithExtra(map[string]any{"id_token": raw})
	}

	user, err := role.GetKeycloakUserInfo(context.Background(), token(idToken("expected")), "expected")
	if err != nil {
		t.Fatal(err)
	}
	if user.Account != "user@cam.ac.uk" || user.DisplayName != "Cambridge User" {
		t.Fatalf("unexpected user information: %+v", user)
	}

	if _, err := role.GetKeycloakUserInfo(context.Background(), token(idToken("wrong")), "expected"); err == nil || !strings.Contains(err.Error(), "nonce") {
		t.Fatalf("expected nonce validation error, got %v", err)
	}

	userinfoSubject.Store("another-user")
	if _, err := role.GetKeycloakUserInfo(context.Background(), token(idToken("expected")), "expected"); err == nil || !strings.Contains(err.Error(), "subject") {
		t.Fatalf("expected subject validation error, got %v", err)
	}
}
