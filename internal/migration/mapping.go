package migration

var targetStatuses = map[SourceStatus]TargetStatus{
	StatusCurrent:   "watching",
	StatusRepeating: "watching",
	StatusPlanning:  "plan_to_watch",
	StatusCompleted: "completed",
	StatusPaused:    "on_hold",
	StatusDropped:   "dropped",
}

func newJob(entry SourceEntry, existing map[MediaRef]TargetUpdate) Job {
	job := Job{Entry: entry, Action: ActionCreate}
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

	if current, found := existing[entry.MediaRef]; found {
		currentCopy := current
		job.Current = &currentCopy
		job.Action = ActionUpdate
		if sameTarget(current, job.Update) {
			job.Action = ActionSkip
			job.Reason = "already up to date"
		}
	}

	return job
}

func sameTarget(current, desired TargetUpdate) bool {
	if current.Status != desired.Status || current.Score != desired.Score || current.Progress != desired.Progress || current.Volumes != desired.Volumes || current.Repeat != desired.Repeat || current.Repeating != desired.Repeating || current.Notes != desired.Notes {
		return false
	}

	if desired.StartDate != "" && current.StartDate != desired.StartDate {
		return false
	}

	if desired.FinishDate != "" && current.FinishDate != desired.FinishDate {
		return false
	}

	return true
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
