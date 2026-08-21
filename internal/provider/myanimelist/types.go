package myanimelist

import "moon-prism-power/internal/migration"

type listResponse struct {
	Data   []listEntry `json:"data"`
	Paging listPaging  `json:"paging"`
}

type detailResponse struct {
	ID          int         `json:"id"`
	NumEpisodes int         `json:"num_episodes"`
	NumChapters int         `json:"num_chapters"`
	ListStatus  *listStatus `json:"my_list_status"`
}

type listEntry struct {
	Node       listNode   `json:"node"`
	ListStatus listStatus `json:"list_status"`
}

type listNode struct {
	ID          int `json:"id"`
	NumEpisodes int `json:"num_episodes"`
	NumChapters int `json:"num_chapters"`
}

type listStatus struct {
	Status             migration.TargetStatus `json:"status"`
	Score              int                    `json:"score"`
	NumEpisodesWatched int                    `json:"num_episodes_watched"`
	NumChaptersRead    int                    `json:"num_chapters_read"`
	NumVolumesRead     int                    `json:"num_volumes_read"`
	NumTimesRewatched  int                    `json:"num_times_rewatched"`
	NumTimesReread     int                    `json:"num_times_reread"`
	IsRewatching       bool                   `json:"is_rewatching"`
	IsRereading        bool                   `json:"is_rereading"`
	Comments           string                 `json:"comments"`
	StartDate          string                 `json:"start_date"`
	FinishDate         string                 `json:"finish_date"`
}

type listPaging struct {
	Next string `json:"next"`
}

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type callbackResult struct {
	Code  string
	State string
	Error string
}
