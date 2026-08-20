package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/joho/godotenv"

	"moon-prism-power/internal/migration"
	"moon-prism-power/internal/platform/browser"
	"moon-prism-power/internal/platform/httpclient"
	"moon-prism-power/internal/provider/anilist"
	"moon-prism-power/internal/provider/mal"
)

func Run(ctx context.Context, input io.Reader, output, errorOutput io.Writer) error {
	if err := loadDotEnv(".env"); err != nil {
		return err
	}

	username := os.Getenv("ANILIST_USER")
	if username == "" {
		return errors.New("ANILIST_USER is required (set it in .env or the environment)")
	}

	clientID := os.Getenv("MAL_CLIENT_ID")
	if clientID == "" {
		return errors.New("MAL_CLIENT_ID is required (set it in .env or the environment)")
	}

	httpClient := httpclient.New()
	malClient := mal.NewClient(httpClient)
	token, err := mal.Authorize(ctx, httpClient, clientID, func(url string) {
		fmt.Fprintf(output, "Open this URL to authorize MyAnimeList access:\n%s\n", url)
		browser.Open(url)
	})

	if err != nil {
		return err
	}

	malClient.SetToken(token)
	service := migration.NewService(anilist.NewClient(httpClient), malClient)
	plan, err := service.Plan(ctx, username)
	if err != nil {
		return err
	}

	fmt.Fprint(output, summary(plan))
	if !confirm(input, output) {
		fmt.Fprintln(output, "Migration canceled. No changes were made.")
		return nil
	}

	result, applyErr := service.Apply(ctx, plan)
	for _, job := range result.Failed {
		fmt.Fprintf(errorOutput, "failed: %s: %s\n", job.Entry.Title, job.Reason)
	}

	fmt.Fprintf(output, "\nMigration complete: %d succeeded, %d skipped, %d failed.\n", result.Succeeded, result.Skipped, len(result.Failed))
	return applyErr
}

func loadDotEnv(path string) error {
	err := godotenv.Load(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}
