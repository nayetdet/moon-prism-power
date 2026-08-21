package mal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"moon-prism-power/internal/utils"
)

const callbackURL = "http://" + callbackAddr + "/callback"

func Authorize(ctx context.Context, client *http.Client, clientID string, openURL func(string)) (OAuthToken, error) {
	if client == nil {
		client = http.DefaultClient
	}

	verifier, err := utils.RandomText(64)
	if err != nil {
		return OAuthToken{}, err
	}

	state, err := utils.RandomText(32)
	if err != nil {
		return OAuthToken{}, err
	}

	callback, server, err := startCallbackServer()
	if err != nil {
		return OAuthToken{}, err
	}

	defer server.Shutdown(context.Background())
	openURL(authorizationURL(clientID, verifier, state))
	select {
	case result := <-callback:
		if result.Error != "" {
			return OAuthToken{}, fmt.Errorf("MAL authorization: %s", result.Error)
		}
		if result.State != state || result.Code == "" {
			return OAuthToken{}, fmt.Errorf("invalid authorization callback")
		}
		return exchange(ctx, client, clientID, result.Code, verifier)
	case <-ctx.Done():
		return OAuthToken{}, ctx.Err()
	}
}

func authorizationURL(clientID, verifier, state string) string {
	return "https://myanimelist.net/v1/oauth2/authorize?" + url.Values{
		"response_type": {"code"}, "client_id": {clientID}, "redirect_uri": {callbackURL},
		"code_challenge": {verifier}, "code_challenge_method": {"plain"}, "state": {state},
	}.Encode()
}

func Refresh(ctx context.Context, client *http.Client, clientID, refreshToken string) (OAuthToken, error) {
	form := url.Values{
		"client_id":     {clientID},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	return requestToken(ctx, client, form)
}

func exchange(ctx context.Context, client *http.Client, clientID, code, verifier string) (OAuthToken, error) {
	form := url.Values{"client_id": {clientID}, "code": {code}, "code_verifier": {verifier}, "grant_type": {"authorization_code"}, "redirect_uri": {callbackURL}}
	return requestToken(ctx, client, form)
}

func requestToken(ctx context.Context, client *http.Client, form url.Values) (OAuthToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://myanimelist.net/v1/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthToken{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return OAuthToken{}, err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body := utils.ReadErrorBody(resp.Body)
		return OAuthToken{}, fmt.Errorf("MAL token: %s: %s", resp.Status, strings.TrimSpace(body))
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return OAuthToken{}, fmt.Errorf("decode MAL token: %w", err)
	}

	if token.AccessToken == "" {
		return OAuthToken{}, fmt.Errorf("MAL token: response has no access_token")
	}

	return token, nil
}
