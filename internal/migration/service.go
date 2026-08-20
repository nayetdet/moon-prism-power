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
	result := Result{}
	for _, job := range plan.Jobs {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		if job.Reason != "" {
			result.Skipped++
			continue
		}

		if err := s.destination.Update(ctx, job.Update); err != nil {
			job.Reason = err.Error()
			result.Failed = append(result.Failed, job)
			continue
		}

		result.Succeeded++
	}

	if len(result.Failed) > 0 {
		return result, fmt.Errorf("%d item(s) could not be migrated", len(result.Failed))
	}

	return result, nil
}
