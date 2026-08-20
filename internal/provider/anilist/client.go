package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"moon-prism-power/internal/migration"
	"moon-prism-power/internal/utils"
)

const endpoint = "https://graphql.anilist.co"
const query = `query ($username: String!, $type: MediaType!) { MediaListCollection(userName: $username, type: $type) { lists { entries { status score progress progressVolumes repeat notes startedAt { year month day } completedAt { year month day } media { idMal title { userPreferred } } } } } }`

type Client struct{ http *http.Client }

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{http: httpClient}
}

func (c *Client) List(ctx context.Context, username string) ([]migration.SourceEntry, error) {
	var all []migration.SourceEntry
	for _, kind := range []migration.MediaKind{migration.Anime, migration.Manga} {
		body, err := json.Marshal(mediaListRequest{Query: query, Variables: mediaListVariables{Username: username, MediaType: string(kind)}})
		if err != nil {
			return nil, fmt.Errorf("encode AniList query: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 300 {
			responseBody := utils.ReadErrorBody(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("AniList returned %s: %s", resp.Status, responseBody)
		}

		var payload mediaListResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode AniList response: %w", err)
		}

		resp.Body.Close()
		if len(payload.Errors) > 0 {
			return nil, fmt.Errorf("AniList: %s", payload.Errors[0].Message)
		}

		all = append(all, payload.entries(kind)...)
	}

	return all, nil
}
