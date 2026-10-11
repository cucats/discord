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
)

var keycloakAdminHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

const (
	discordIDAttribute       = "discord_id"
	discordUsernameAttribute = "discord_username"
	collegeAttribute         = "college"
)

// StoreVerificationInKeycloak saves the result against the subject of the
// verified Keycloak ID token. The service account is used only for this write.
func StoreVerificationInKeycloak(ctx context.Context, keycloakUserID, discordID, discordUsername string, college College) error {
	if keycloakUserID == "" || discordID == "" || discordUsername == "" {
		return errors.New("verification identity is incomplete")
	}
	adminURL, err := keycloakUserURL(config.KeycloakIssuer, keycloakUserID)
	if err != nil {
		return err
	}
	token, err := keycloakWriterToken(ctx)
	if err != nil {
		return err
	}

	user, err := readKeycloakUser(ctx, adminURL, token)
	if err != nil {
		return err
	}
	attributes := map[string][]string{}
	if raw := user["attributes"]; len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &attributes); err != nil {
			return fmt.Errorf("decode Keycloak user attributes: %w", err)
		}
	}
	attributes[discordIDAttribute] = []string{discordID}
	attributes[discordUsernameAttribute] = []string{discordUsername}
	if collegeName := college.Name(); collegeName != "" {
		attributes[collegeAttribute] = []string{collegeName}
	} else {
		delete(attributes, collegeAttribute)
	}
	rawAttributes, err := json.Marshal(attributes)
	if err != nil {
		return err
	}
	user["attributes"] = rawAttributes
	body, err := json.Marshal(user)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, adminURL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := keycloakAdminHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("update Keycloak user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("update Keycloak user: HTTP %d", resp.StatusCode)
	}

	// Keycloak may accept a user update while ignoring attributes that are not
	// allowed by the realm's user profile. Read them back before reporting success.
	storedUser, err := readKeycloakUser(ctx, adminURL, token)
	if err != nil {
		return err
	}
	var storedAttributes map[string][]string
	if err := json.Unmarshal(storedUser["attributes"], &storedAttributes); err != nil {
		return fmt.Errorf("decode saved Keycloak attributes: %w", err)
	}
	if first(storedAttributes[discordIDAttribute]) != discordID ||
		first(storedAttributes[discordUsernameAttribute]) != discordUsername ||
		first(storedAttributes[collegeAttribute]) != college.Name() {
		return errors.New("Keycloak did not retain the verification attributes")
	}
	return nil
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func keycloakUserURL(issuer, userID string) (string, error) {
	u, err := url.Parse(strings.TrimRight(issuer, "/"))
	if err != nil || u.Scheme != "https" && u.Scheme != "http" || u.Host == "" {
		return "", errors.New("invalid Keycloak issuer URL")
	}
	basePath, realm, ok := strings.Cut(u.Path, "/realms/")
	if !ok || realm == "" || strings.Contains(realm, "/") {
		return "", errors.New("Keycloak issuer URL does not contain one realm")
	}
	return u.Scheme + "://" + u.Host + basePath + "/admin/realms/" + url.PathEscape(realm) + "/users/" + url.PathEscape(userID), nil
}

func keycloakWriterToken(ctx context.Context) (string, error) {
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {config.KeycloakWriterClientID},
		"client_secret": {config.KeycloakWriterClientSecret},
	}
	tokenURL := strings.TrimRight(config.KeycloakIssuer, "/") + "/protocol/openid-connect/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := keycloakAdminHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request Keycloak writer token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request Keycloak writer token: HTTP %d", resp.StatusCode)
	}
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", fmt.Errorf("decode Keycloak writer token: %w", err)
	}
	if result.AccessToken == "" {
		return "", errors.New("Keycloak writer token is missing")
	}
	return result.AccessToken, nil
}

func readKeycloakUser(ctx context.Context, adminURL, token string) (map[string]json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, adminURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := keycloakAdminHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("read Keycloak user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("read Keycloak user: HTTP %d", resp.StatusCode)
	}
	var user map[string]json.RawMessage
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode Keycloak user: %w", err)
	}
	return user, nil
}
