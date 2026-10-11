package role

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"cucats.org/discord/config"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	KeycloakOAuth    *oauth2.Config
	keycloakProvider *oidc.Provider
	keycloakVerifier *oidc.IDTokenVerifier
)

func InitKeycloakOAuth(ctx context.Context) error {
	discoveryCtx := oidc.ClientContext(ctx, &http.Client{Timeout: 10 * time.Second})
	var provider *oidc.Provider
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		provider, err = oidc.NewProvider(discoveryCtx, config.KeycloakIssuer)
		if err == nil || ctx.Err() != nil {
			break
		}
		slog.Warn("Keycloak discovery attempt failed", "attempt", attempt, "error", err)
	}
	if err != nil {
		return fmt.Errorf("discover Keycloak issuer: %w", err)
	}

	keycloakProvider = provider
	keycloakVerifier = provider.Verifier(&oidc.Config{ClientID: config.KeycloakClientID})
	KeycloakOAuth = &oauth2.Config{
		ClientID:     config.KeycloakClientID,
		ClientSecret: config.KeycloakClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  config.Host + "/keycloak/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return nil
}

type keycloakClaims struct {
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
}

func GetKeycloakUserInfo(ctx context.Context, token *oauth2.Token, expectedNonce string) (*UserInfo, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, errors.New("Keycloak did not return an ID token")
	}

	idToken, err := keycloakVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verify Keycloak ID token: %w", err)
	}
	if idToken.Nonce != expectedNonce {
		return nil, errors.New("Keycloak ID token nonce does not match")
	}

	var claims keycloakClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("read Keycloak ID token claims: %w", err)
	}

	profile, err := keycloakProvider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("fetch Keycloak userinfo: %w", err)
	}
	if profile.Subject != idToken.Subject {
		return nil, errors.New("Keycloak userinfo subject does not match ID token")
	}

	var profileClaims keycloakClaims
	if err := profile.Claims(&profileClaims); err != nil {
		return nil, fmt.Errorf("read Keycloak userinfo claims: %w", err)
	}
	if profileClaims.Email != "" {
		claims.Email = profileClaims.Email
	}
	if profileClaims.PreferredUsername != "" {
		claims.PreferredUsername = profileClaims.PreferredUsername
	}
	if profileClaims.Name != "" {
		claims.Name = profileClaims.Name
	}
	account := claims.Email
	if account == "" {
		account = claims.PreferredUsername
	}
	if account == "" {
		account = idToken.Subject
	}
	displayName := claims.Name
	if displayName == "" {
		displayName = account
	}

	return &UserInfo{Account: account, DisplayName: displayName}, nil
}
