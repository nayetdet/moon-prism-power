package myanimelist

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

func (c *Client) List(ctx context.Context) (map[migration.MediaRef]migration.TargetUpdate, error) {
	items := map[migration.MediaRef]migration.TargetUpdate{}
	for _, kind := range []migration.MediaKind{migration.Anime, migration.Manga} {
		field := "num_episodes"
		if kind == migration.Manga {
			field = "num_chapters"
		}

		next := apiURL + "/users/@me/" + string(kind) + "list?limit=1000&fields=list_status," + field
		for next != "" {
			req, err := c.newRequest(ctx, http.MethodGet, next, nil)
			if err != nil {
				return nil, err
			}

			req.Header.Set("Cache-Control", "no-cache")
			req.Header.Set("Pragma", "no-cache")
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
				items[migration.MediaRef{Kind: kind, MALID: entry.Node.ID}] = toTargetUpdate(kind, entry.Node, entry.ListStatus)
			}

			next = page.Paging.Next
		}
	}

	return items, nil
}

func (c *Client) Get(ctx context.Context, ref migration.MediaRef) (migration.TargetUpdate, bool, error) {
	field := "num_episodes"
	if ref.Kind == migration.Manga {
		field = "num_chapters"
	}

	endpoint := fmt.Sprintf("%s/%s/%d?fields=my_list_status,%s", apiURL, ref.Kind, ref.MALID, field)
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return migration.TargetUpdate{}, false, err
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := c.http.Do(req)
	if err != nil {
		return migration.TargetUpdate{}, false, err
	}

	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		body := utils.ReadErrorBody(resp.Body)
		return migration.TargetUpdate{}, false, fmt.Errorf("MAL item: %s: %s", resp.Status, strings.TrimSpace(body))
	}

	var details detailResponse
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return migration.TargetUpdate{}, false, fmt.Errorf("decode MAL item: %w", err)
	}

	if details.ListStatus == nil {
		return migration.TargetUpdate{}, false, nil
	}

	return toTargetUpdate(ref.Kind, listNode{
		ID: ref.MALID, NumEpisodes: details.NumEpisodes, NumChapters: details.NumChapters,
	}, *details.ListStatus), true, nil
}

func toTargetUpdate(kind migration.MediaKind, node listNode, status listStatus) migration.TargetUpdate {
	progress := status.NumEpisodesWatched
	repeat := status.NumTimesRewatched
	repeating := status.IsRewatching
	if kind == migration.Manga {
		progress = status.NumChaptersRead
		repeat = status.NumTimesReread
		repeating = status.IsRereading
	}

	return migration.TargetUpdate{
		MediaRef:   migration.MediaRef{Kind: kind, MALID: node.ID},
		Status:     status.Status,
		Score:      status.Score,
		Progress:   progress,
		Volumes:    status.NumVolumesRead,
		Repeat:     repeat,
		Repeating:  repeating,
		Notes:      status.Comments,
		StartDate:  status.StartDate,
		FinishDate: status.FinishDate,
		ProgressLimit: func() int {
			if kind == migration.Anime {
				return node.NumEpisodes
			}

			return node.NumChapters
		}(),
	}
}

func (c *Client) Update(ctx context.Context, item migration.TargetUpdate) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("wait for MAL rate limit: %w", err)
	}

	form := url.Values{"status": {string(item.Status)}, "score": {fmt.Sprint(item.Score)}}
	if item.Kind == migration.Anime {
		form.Set("num_watched_episodes", fmt.Sprint(item.Progress))
	} else {
		form.Set("num_chapters_read", fmt.Sprint(item.Progress))
		form.Set("num_volumes_read", fmt.Sprint(item.Volumes))
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

	var updated listStatus
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return fmt.Errorf("decode MAL update: %w", err)
	}

	if updated.Status != item.Status {
		return fmt.Errorf("MAL update returned status %q, expected %q", updated.Status, item.Status)
	}

	if item.Kind == migration.Anime && updated.NumEpisodesWatched != item.Progress {
		return fmt.Errorf("MAL update returned progress %d, expected %d", updated.NumEpisodesWatched, item.Progress)
	}

	if item.Kind == migration.Manga && (updated.NumChaptersRead != item.Progress || updated.NumVolumesRead != item.Volumes) {
		return fmt.Errorf("MAL update returned progress %d/%d, expected %d/%d", updated.NumChaptersRead, updated.NumVolumesRead, item.Progress, item.Volumes)
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
