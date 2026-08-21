package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"moon-prism-power/internal/migration"
	"moon-prism-power/internal/platform/httpclient"
	"moon-prism-power/internal/provider/anilist"
	"moon-prism-power/internal/provider/myanimelist"
	"moon-prism-power/internal/report"
)

const reportDirectory = "data/report"

func RunMPP(ctx context.Context, input io.Reader, output, errorOutput io.Writer) error {
	if err := loadDotEnv(".env"); err != nil {
		return err
	}

	username, err := requiredEnv("ANILIST_USER", "set it in .env or the environment")
	if err != nil {
		return err
	}

	refreshToken := os.Getenv("MAL_REFRESH_TOKEN")
	if refreshToken == "" {
		return fmt.Errorf("MAL_REFRESH_TOKEN is required (run `mpp-auth` and set it in .env or the environment)")
	}

	httpClient := httpclient.New()
	clientID, clientErr := requiredEnv("MAL_CLIENT_ID", "set it in .env or the environment")
	if clientErr != nil {
		return clientErr
	}

	token, refreshErr := myanimelist.Refresh(ctx, httpClient, clientID, refreshToken)
	if refreshErr != nil {
		return refreshErr
	}

	myanimelistClient := myanimelist.NewClient(httpClient)
	myanimelistClient.SetToken(token.AccessToken)
	service := migration.NewService(anilist.NewClient(httpClient), myanimelistClient)
	plan, err := service.Plan(ctx, username)
	if err != nil {
		return err
	}

	startedAt := time.Now().UTC()
	executionReport := report.New(plan, startedAt)
	reportPath := filepath.Join(reportDirectory, "migration-"+executionReport.ExecutionID+".json")
	if err := executionReport.Write(reportPath); err != nil {
		return err
	}

	_, _ = fmt.Fprint(output, planSummary(plan))
	if os.Getenv("AUTO_CONFIRM") != "true" && !confirm(input, output) {
		executionReport.SetCanceled(time.Now())
		if err := executionReport.Write(reportPath); err != nil {
			return err
		}

		_, _ = fmt.Fprintln(output, "Migration canceled. No changes were made.")
		return nil
	}

	result, applyErr := service.Apply(ctx, plan)
	executionReport.SetResult(result, time.Now(), applyErr)
	reportErr := executionReport.Write(reportPath)
	for _, job := range result.Failed {
		_, _ = fmt.Fprintf(errorOutput, "failed: %s: %s\n", job.Entry.Title, job.Reason)
	}

	_, _ = fmt.Fprintf(output, "\nMigration complete: %d succeeded, %d skipped, %d failed.\n", result.Succeeded, result.Skipped, len(result.Failed))
	return errors.Join(applyErr, reportErr)
}
