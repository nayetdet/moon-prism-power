package migration

import "context"

type MediaKind string

const (
	Anime MediaKind = "anime"
	Manga MediaKind = "manga"
)

type SourceStatus string

const (
	StatusCurrent   SourceStatus = "CURRENT"
	StatusRepeating SourceStatus = "REPEATING"
	StatusPlanning  SourceStatus = "PLANNING"
	StatusCompleted SourceStatus = "COMPLETED"
	StatusPaused    SourceStatus = "PAUSED"
	StatusDropped   SourceStatus = "DROPPED"
)

type TargetStatus string

type MediaRef struct {
	Kind  MediaKind
	MALID int
}

type SourceEntry struct {
	MediaRef
	Title, Notes, StartDate, FinishDate string
	Status                              SourceStatus
	Score, Progress, Volumes, Repeat    int
}

type TargetUpdate struct {
	MediaRef
	Status                           TargetStatus
	Notes, StartDate, FinishDate     string
	Score, Progress, Volumes, Repeat int
	Repeating                        bool
}

type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionSkip   Action = "skip"
)

type Source interface {
	List(context.Context, string) ([]SourceEntry, error)
}

type Destination interface {
	List(context.Context) (map[MediaRef]TargetUpdate, error)
	Update(context.Context, TargetUpdate) error
}

type Job struct {
	Entry  SourceEntry
	Update TargetUpdate
	Action Action
	Reason string
}

type Plan struct {
	AniListUser string
	Jobs        []Job
}

type Result struct {
	Succeeded int
	Skipped   int
	Failed    []Job
	Jobs      []JobResult
}

type JobStatus string

const (
	JobPending     JobStatus = "pending"
	JobSkipped     JobStatus = "skipped"
	JobSucceeded   JobStatus = "succeeded"
	JobFailed      JobStatus = "failed"
	JobInterrupted JobStatus = "interrupted"
)

type JobResult struct {
	Status JobStatus
	Reason string
}
