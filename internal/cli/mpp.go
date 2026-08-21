package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"moon-prism-power/internal/migration"
	"moon-prism-power/internal/platform/httpclient"
	"moon-prism-power/internal/provider/anilist"
	"moon-prism-power/internal/provider/mal"
)

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

	token, refreshErr := mal.Refresh(ctx, httpClient, clientID, refreshToken)
	if refreshErr != nil {
		return refreshErr
	}

	malClient := mal.NewClient(httpClient)
	malClient.SetToken(token.AccessToken)
	service := migration.NewService(anilist.NewClient(httpClient), malClient)
	plan, err := service.Plan(ctx, username)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprint(output, summary(plan))
	if !confirm(input, output) {
		_, _ = fmt.Fprintln(output, "Migration canceled. No changes were made.")
		return nil
	}

	result, applyErr := service.Apply(ctx, plan)
	for _, job := range result.Failed {
		_, _ = fmt.Fprintf(errorOutput, "failed: %s: %s\n", job.Entry.Title, job.Reason)
	}

	_, _ = fmt.Fprintf(output, "\nMigration complete: %d succeeded, %d skipped, %d failed.\n", result.Succeeded, result.Skipped, len(result.Failed))
	return applyErr
}
