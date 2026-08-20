package migration

var targetStatuses = map[SourceStatus]TargetStatus{
	StatusCurrent:   "watching",
	StatusRepeating: "watching",
	StatusPlanning:  "plan_to_watch",
	StatusCompleted: "completed",
	StatusPaused:    "on_hold",
	StatusDropped:   "dropped",
}

func newJob(entry SourceEntry, existing map[MediaRef]struct{}) Job {
	job := Job{Entry: entry, Action: ActionCreate}
	if _, found := existing[entry.MediaRef]; found {
		job.Action = ActionUpdate
	}

	if entry.MALID == 0 {
		job.Action, job.Reason = ActionSkip, "AniList did not provide a MyAnimeList ID"
		return job
	}

	status := toTargetStatus(entry.Kind, entry.Status)
	if status == "" {
		job.Action, job.Reason = ActionSkip, "unsupported AniList status: "+string(entry.Status)
		return job
	}

	job.Update = TargetUpdate{
		MediaRef: entry.MediaRef, Status: status, Score: toTargetScore(entry.Score),
		Progress: entry.Progress, Volumes: entry.Volumes, Repeat: entry.Repeat,
		Repeating: entry.Status == StatusRepeating, Notes: entry.Notes,
		StartDate: entry.StartDate, FinishDate: entry.FinishDate,
	}

	return job
}

func toTargetStatus(kind MediaKind, status SourceStatus) TargetStatus {
	if kind != Anime && kind != Manga {
		return ""
	}

	targetStatus := targetStatuses[status]
	if targetStatus == "" || kind != Manga {
		return targetStatus
	}

	if targetStatus == "watching" {
		return "reading"
	}

	if targetStatus == "plan_to_watch" {
		return "plan_to_read"
	}

	return targetStatus
}

func toTargetScore(value int) int {
	return min((max(value, 0)+5)/10, 10)
}
