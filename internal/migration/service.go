package migration

import (
	"context"
	"fmt"
)

type Service struct {
	source      Source
	destination Destination
}

func NewService(source Source, destination Destination) Service {
	return Service{source: source, destination: destination}
}

func (s Service) Plan(ctx context.Context, username string) (Plan, error) {
	entries, err := s.source.List(ctx, username)
	if err != nil {
		return Plan{}, err
	}

	existing, err := s.destination.List(ctx)
	if err != nil {
		return Plan{}, err
	}

	jobs := make([]Job, 0, len(entries))
	for _, entry := range entries {
		jobs = append(jobs, newJob(entry, existing))
	}

	return Plan{AniListUser: username, Jobs: jobs}, nil
}

func (s Service) Apply(ctx context.Context, plan Plan) (Result, error) {
	result := Result{Jobs: make([]JobResult, len(plan.Jobs))}
	for i, job := range plan.Jobs {
		if err := ctx.Err(); err != nil {
			for j := i; j < len(result.Jobs); j++ {
				result.Jobs[j] = JobResult{Status: JobInterrupted, Reason: err.Error()}
			}
			return result, err
		}

		if job.Reason != "" {
			result.Jobs[i] = JobResult{Status: JobSkipped, Reason: job.Reason}
			result.Skipped++
			continue
		}

		if err := s.destination.Update(ctx, job.Update); err != nil {
			job.Reason = err.Error()
			result.Jobs[i] = JobResult{Status: JobFailed, Reason: job.Reason}
			result.Failed = append(result.Failed, job)
			continue
		}

		result.Jobs[i] = JobResult{Status: JobSucceeded}
		result.Succeeded++
	}

	if len(result.Failed) > 0 {
		return result, fmt.Errorf("%d item(s) could not be migrated", len(result.Failed))
	}

	return result, nil
}
