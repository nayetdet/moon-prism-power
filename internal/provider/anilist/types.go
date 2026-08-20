package anilist

import (
	"fmt"

	"moon-prism-power/internal/migration"
)

type mediaListRequest struct {
	Query     string             `json:"query"`
	Variables mediaListVariables `json:"variables"`
}

type mediaListVariables struct {
	Username  string `json:"username"`
	MediaType string `json:"type"`
}

type mediaListResponse struct {
	Data   mediaListData `json:"data"`
	Errors []apiError    `json:"errors"`
}

type mediaListData struct {
	Collection mediaListCollection `json:"MediaListCollection"`
}

type mediaListCollection struct {
	Lists []mediaList `json:"lists"`
}

type mediaList struct {
	Entries []mediaListEntry `json:"entries"`
}

type mediaListEntry struct {
	Status          migration.SourceStatus `json:"status"`
	Notes           string                 `json:"notes"`
	Score           int                    `json:"score"`
	Progress        int                    `json:"progress"`
	ProgressVolumes int                    `json:"progressVolumes"`
	Repeat          int                    `json:"repeat"`
	StartedAt       apiDate                `json:"startedAt"`
	CompletedAt     apiDate                `json:"completedAt"`
	Media           anilistMedia           `json:"media"`
}

type anilistMedia struct {
	MALID int        `json:"idMal"`
	Title mediaTitle `json:"title"`
}

type mediaTitle struct {
	UserPreferred string `json:"userPreferred"`
}

type apiError struct {
	Message string `json:"message"`
}

type apiDate struct {
	Year  int
	Month int
	Day   int
}

func (d apiDate) String() string {
	if d.Year == 0 || d.Month == 0 || d.Day == 0 {
		return ""
	}

	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (r mediaListResponse) entries(kind migration.MediaKind) []migration.SourceEntry {
	var entries []migration.SourceEntry
	for _, list := range r.Data.Collection.Lists {
		for _, item := range list.Entries {
			entries = append(entries, migration.SourceEntry{
				MediaRef: migration.MediaRef{Kind: kind, MALID: item.Media.MALID},
				Title:    item.Media.Title.UserPreferred, Status: item.Status, Score: item.Score,
				Progress: item.Progress, Volumes: item.ProgressVolumes, Repeat: item.Repeat,
				Notes: item.Notes, StartDate: item.StartedAt.String(), FinishDate: item.CompletedAt.String(),
			})
		}
	}

	return entries
}
