package mal

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"moon-prism-power/internal/utils"
)

const callbackURL = "http://127.0.0.1:8787/callback"

func Authorize(ctx context.Context, client *http.Client, clientID string, openURL func(string)) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}

	verifier, err := utils.RandomText(64)
	if err != nil {
		return "", err
	}

	state, err := utils.RandomText(32)
	if err != nil {
		return "", err
	}

	callback, server, err := startCallbackServer()
	if err != nil {
		return "", err
	}

	defer server.Shutdown(context.Background())
	openURL(authorizationURL(clientID, verifier, state))
	select {
	case result := <-callback:
		if result.Error != "" {
			return "", fmt.Errorf("MAL authorization: %s", result.Error)
		}
		if result.State != state || result.Code == "" {
			return "", fmt.Errorf("invalid authorization callback")
		}
		return exchange(ctx, client, clientID, result.Code, verifier)
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func authorizationURL(clientID, verifier, state string) string {
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return "https://myanimelist.net/v1/oauth2/authorize?" + url.Values{
		"response_type": {"code"}, "client_id": {clientID}, "redirect_uri": {callbackURL},
		"code_challenge": {challenge}, "code_challenge_method": {"S256"}, "state": {state},
	}.Encode()
}

func exchange(ctx context.Context, client *http.Client, clientID, code, verifier string) (string, error) {
	form := url.Values{"client_id": {clientID}, "code": {code}, "code_verifier": {verifier}, "grant_type": {"authorization_code"}, "redirect_uri": {callbackURL}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://myanimelist.net/v1/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body := utils.ReadErrorBody(resp.Body)
		return "", fmt.Errorf("MAL token: %s: %s", resp.Status, strings.TrimSpace(body))
	}

	var token tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", fmt.Errorf("decode MAL token: %w", err)
	}

	if token.AccessToken == "" {
		return "", fmt.Errorf("MAL token: response has no access_token")
	}

	return token.AccessToken, nil
}
