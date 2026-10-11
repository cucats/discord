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
)

const graphBaseURL = "https://graph.microsoft.com/v1.0"

var ErrGraphUnauthorized = errors.New("Microsoft Graph rejected the Entra token")

var graphHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func GetGraphUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	return getGraphUserInfo(ctx, graphHTTPClient, graphBaseURL, accessToken)
}

func getGraphUserInfo(ctx context.Context, client *http.Client, baseURL, accessToken string) (*UserInfo, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	var profile struct {
		UserPrincipalName string `json:"userPrincipalName"`
		DisplayName       string `json:"displayName"`
	}
	if err := getGraphJSON(ctx, client, baseURL+"/me", accessToken, &profile); err != nil {
		return nil, err
	}

	user := &UserInfo{Account: profile.UserPrincipalName, DisplayName: profile.DisplayName}
	next := baseURL + "/me/memberOf"
	for page := 0; next != ""; page++ {
		if page >= 100 {
			return nil, fmt.Errorf("Microsoft Graph returned too many membership pages")
		}
		if err := validateGraphPageURL(baseURL, next); err != nil {
			return nil, err
		}
		var membership struct {
			Value []struct {
				ID string `json:"id"`
			} `json:"value"`
			Next string `json:"@odata.nextLink"`
		}
		if err := getGraphJSON(ctx, client, next, accessToken, &membership); err != nil {
			return nil, err
		}
		groups := make([]string, 0, len(membership.Value))
		for _, member := range membership.Value {
			groups = append(groups, member.ID)
		}
		addCambridgeGroups(user, groups)
		next = membership.Next
	}
	return user, nil
}

func validateGraphPageURL(baseURL, pageURL string) error {
	base, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	page, err := url.Parse(pageURL)
	if err != nil {
		return err
	}
	if page.Scheme != base.Scheme || page.Host != base.Host || page.Path != base.Path+"/me/memberOf" {
		return fmt.Errorf("Microsoft Graph returned an unexpected membership page URL")
	}
	return nil
}

func getGraphJSON(ctx context.Context, client *http.Client, endpoint, accessToken string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request Microsoft Graph: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrGraphUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Microsoft Graph returned HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(out); err != nil {
		return fmt.Errorf("decode Microsoft Graph response: %w", err)
	}
	return nil
}
