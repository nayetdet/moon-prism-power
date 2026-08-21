package mal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/time/rate"

	"moon-prism-power/internal/migration"
	"moon-prism-power/internal/utils"
)

const apiURL = "https://api.myanimelist.net/v2"

type Client struct {
	http    *http.Client
	token   string
	limiter *rate.Limiter
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{http: httpClient, limiter: rate.NewLimiter(2, 1)}
}

func (c *Client) SetToken(token string) { c.token = token }

func (c *Client) List(ctx context.Context) (map[migration.MediaRef]struct{}, error) {
	items := map[migration.MediaRef]struct{}{}
	for _, kind := range []migration.MediaKind{migration.Anime, migration.Manga} {
		next := apiURL + "/users/@me/" + string(kind) + "list?limit=1000"
		for next != "" {
			req, err := c.newRequest(ctx, http.MethodGet, next, nil)
			if err != nil {
				return nil, err
			}

			resp, err := c.http.Do(req)
			if err != nil {
				return nil, err
			}

			if resp.StatusCode >= 300 {
				body := utils.ReadErrorBody(resp.Body)
				_ = resp.Body.Close()
				return nil, fmt.Errorf("MAL list: %s: %s", resp.Status, strings.TrimSpace(body))
			}

			var page listResponse
			err = json.NewDecoder(resp.Body).Decode(&page)
			_ = resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("decode MAL list: %w", err)
			}

			for _, entry := range page.Data {
				items[migration.MediaRef{Kind: kind, MALID: entry.Node.ID}] = struct{}{}
			}

			next = page.Paging.Next
		}
	}

	return items, nil
}

func (c *Client) Update(ctx context.Context, item migration.TargetUpdate) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("wait for MAL rate limit: %w", err)
	}

	form := url.Values{"status": {string(item.Status)}, "score": {fmt.Sprint(item.Score)}, "num_times_rewatched": {fmt.Sprint(item.Repeat)}, "comments": {item.Notes}}
	if item.Kind == migration.Anime {
		form.Set("num_watched_episodes", fmt.Sprint(item.Progress))
		form.Set("is_rewatching", fmt.Sprint(item.Repeating))
	} else {
		form.Set("num_chapters_read", fmt.Sprint(item.Progress))
		form.Set("num_volumes_read", fmt.Sprint(item.Volumes))
		form.Set("is_rereading", fmt.Sprint(item.Repeating))
	}

	if item.StartDate != "" {
		form.Set("start_date", item.StartDate)
	}

	if item.FinishDate != "" {
		form.Set("finish_date", item.FinishDate)
	}

	req, err := c.newRequest(ctx, http.MethodPatch, fmt.Sprintf("%s/%s/%d/my_list_status", apiURL, item.Kind, item.MALID), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		body := utils.ReadErrorBody(resp.Body)
		return fmt.Errorf("MAL returned %s: %s", resp.Status, strings.TrimSpace(body))
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err == nil {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, err
}
