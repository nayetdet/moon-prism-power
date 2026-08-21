package report

import (
	"context"
	"errors"
	"time"

	"moon-prism-power/internal/migration"
)

func New(plan migration.Plan, startedAt time.Time) Report {
	report := Report{
		SchemaVersion: SchemaVersion,
		ExecutionID:   startedAt.UTC().Format("20060102T150405.000000000Z"),
		Status:        "planned",
		StartedAt:     startedAt.UTC(),
		AniListUser:   plan.AniListUser,
		Jobs:          make([]JobReport, len(plan.Jobs)),
	}

	for i, job := range plan.Jobs {
		status := migration.JobPending
		if job.Reason != "" {
			status = migration.JobSkipped
		}

		report.Jobs[i] = jobReport(job, migration.JobResult{Status: status, Reason: job.Reason})
		switch job.Action {
		case migration.ActionCreate:
			report.Summary.Create++
		case migration.ActionUpdate:
			report.Summary.Update++
		case migration.ActionSkip:
			report.Summary.Skip++
		}
	}

	return report
}

func (r *Report) SetCanceled(now time.Time) {
	finishedAt := now.UTC()
	r.Status = "canceled"
	r.FinishedAt = &finishedAt

	for i := range r.Jobs {
		if r.Jobs[i].Status == migration.JobPending {
			r.Jobs[i].Status = migration.JobSkipped
			r.Jobs[i].Reason = "migration canceled"
		}
	}
}

func (r *Report) SetResult(result migration.Result, now time.Time, applyErr error) {
	finishedAt := now.UTC()
	r.FinishedAt = &finishedAt
	r.Summary.Succeeded = result.Succeeded
	r.Summary.Skipped = result.Skipped
	r.Summary.Failed = len(result.Failed)

	switch {
	case errors.Is(applyErr, context.Canceled), errors.Is(applyErr, context.DeadlineExceeded):
		r.Status = "interrupted"
	case applyErr != nil:
		r.Status = "failed"
		r.Error = applyErr.Error()
	default:
		r.Status = "completed"
	}

	for i, outcome := range result.Jobs {
		if i < len(r.Jobs) {
			r.Jobs[i].Status = outcome.Status
			r.Jobs[i].Reason = outcome.Reason
		}
	}
}

func jobReport(job migration.Job, outcome migration.JobResult) JobReport {
	return JobReport{
		Title: job.Entry.Title, Kind: job.Entry.Kind, MALID: job.Entry.MALID,
		Action: job.Action, Status: outcome.Status, Reason: outcome.Reason,
		Source: Source{
			Title: job.Entry.Title, Status: job.Entry.Status, Score: job.Entry.Score,
			Progress: job.Entry.Progress, Volumes: job.Entry.Volumes, Repeat: job.Entry.Repeat,
			Notes: job.Entry.Notes, StartDate: job.Entry.StartDate, FinishDate: job.Entry.FinishDate,
		},
		Target: Target{
			Status: job.Update.Status, Score: job.Update.Score, Progress: job.Update.Progress,
			Volumes: job.Update.Volumes, Repeat: job.Update.Repeat, Repeating: job.Update.Repeating,
			Notes: job.Update.Notes, StartDate: job.Update.StartDate, FinishDate: job.Update.FinishDate,
		},
	}
}
