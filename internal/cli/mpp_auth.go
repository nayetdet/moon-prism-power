package cli

import (
	"context"
	"fmt"
	"io"

	"moon-prism-power/internal/platform/browser"
	"moon-prism-power/internal/platform/httpclient"
	"moon-prism-power/internal/provider/mal"
)

func RunMPPAuth(ctx context.Context, output io.Writer) error {
	if err := loadDotEnv(".env"); err != nil {
		return err
	}

	clientID, err := requiredEnv("MAL_CLIENT_ID", "set it in .env or the environment")
	if err != nil {
		return err
	}

	httpClient := httpclient.New()
	token, err := mal.Authorize(ctx, httpClient, clientID, func(url string) {
		fmt.Fprintf(output, "Open this URL to authorize MyAnimeList access:\n%s\n", url)
		browser.Open(url)
	})

	if err != nil {
		return err
	}

	fmt.Fprintf(output, "MAL_REFRESH_TOKEN=%s\n", token.RefreshToken)
	return nil
}
