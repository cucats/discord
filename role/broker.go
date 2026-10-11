package role

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cucats.org/discord/config"
	"golang.org/x/oauth2"
)

var brokerHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// GetBrokeredEntraToken uses the client-scoped V2 broker endpoint when enabled,
// falling back to V1 for older Keycloak deployments. The token is associated
// with the authenticated Keycloak user and is never placed in our session or
// cookie.
func GetBrokeredEntraToken(ctx context.Context, keycloakToken *oauth2.Token) (string, error) {
	endpoint := strings.TrimRight(config.KeycloakIssuer, "/") + "/broker/" + url.PathEscape(config.KeycloakEntraAlias) + "/token"
	token, postStatus, err := requestBrokerToken(ctx, http.MethodPost, endpoint, keycloakToken.AccessToken)
	if err != nil {
		return "", err
	}
	if token != "" {
		return token, nil
	}
	token, getStatus, err := requestBrokerToken(ctx, http.MethodGet, endpoint, keycloakToken.AccessToken)
	if err != nil {
		return "", err
	}
	if token != "" {
		return token, nil
	}
	return "", fmt.Errorf("Keycloak did not provide an Entra token (broker V1 HTTP %d, V2 HTTP %d); check Store Tokens and broker token permissions", getStatus, postStatus)
}

func requestBrokerToken(ctx context.Context, method, endpoint, keycloakAccessToken string) (string, int, error) {
	var body io.Reader
	if method == http.MethodPost {
		form := url.Values{
			"client_id":     {config.KeycloakClientID},
			"client_secret": {config.KeycloakClientSecret},
			"token":         {keycloakAccessToken},
		}
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return "", 0, fmt.Errorf("create Keycloak broker request: %w", err)
	}
	if method == http.MethodGet {
		req.Header.Set("Authorization", "Bearer "+keycloakAccessToken)
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Accept", "application/json")
	resp, err := brokerHTTPClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("request Keycloak broker token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode, nil
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", resp.StatusCode, fmt.Errorf("decode Keycloak broker token response: %w", err)
	}
	if result.AccessToken == "" {
		return "", resp.StatusCode, errors.New("Keycloak broker response did not contain an Entra access token")
	}
	return result.AccessToken, resp.StatusCode, nil
}
