package report

import (
	"time"

	"moon-prism-power/internal/migration"
)

const SchemaVersion = 1

type Report struct {
	SchemaVersion int         `json:"schema_version"`
	ExecutionID   string      `json:"execution_id"`
	Status        string      `json:"status"`
	StartedAt     time.Time   `json:"started_at"`
	FinishedAt    *time.Time  `json:"finished_at,omitempty"`
	AniListUser   string      `json:"anilist_user"`
	Summary       Summary     `json:"summary"`
	Jobs          []JobReport `json:"jobs"`
	Error         string      `json:"error,omitempty"`
}

type Summary struct {
	Create    int `json:"create"`
	Update    int `json:"update"`
	Skip      int `json:"skip"`
	Succeeded int `json:"succeeded"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

type JobReport struct {
	Title   string              `json:"title"`
	Kind    migration.MediaKind `json:"kind"`
	MALID   int                 `json:"mal_id"`
	Action  migration.Action    `json:"action"`
	Status  migration.JobStatus `json:"status"`
	Reason  string              `json:"reason,omitempty"`
	Current *Target             `json:"current,omitempty"`
	Source  Source              `json:"source"`
	Target  Target              `json:"target"`
}

type Source struct {
	Title      string                 `json:"title"`
	Status     migration.SourceStatus `json:"status"`
	Score      int                    `json:"score"`
	Progress   int                    `json:"progress"`
	Volumes    int                    `json:"volumes"`
	Repeat     int                    `json:"repeat"`
	Notes      string                 `json:"notes"`
	StartDate  string                 `json:"start_date"`
	FinishDate string                 `json:"finish_date"`
}

type Target struct {
	Status     migration.TargetStatus `json:"status,omitempty"`
	Score      int                    `json:"score,omitempty"`
	Progress   int                    `json:"progress,omitempty"`
	Volumes    int                    `json:"volumes,omitempty"`
	Repeat     int                    `json:"repeat,omitempty"`
	Repeating  bool                   `json:"repeating,omitempty"`
	Notes      string                 `json:"notes,omitempty"`
	StartDate  string                 `json:"start_date,omitempty"`
	FinishDate string                 `json:"finish_date,omitempty"`
}
