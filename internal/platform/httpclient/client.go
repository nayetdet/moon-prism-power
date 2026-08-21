package httpclient

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const DefaultTimeout = 2 * time.Minute

func New() *http.Client {
	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = 3
	client.RetryWaitMin = 250 * time.Millisecond
	client.RetryWaitMax = 4 * time.Second
	client.Backoff = retryablehttp.RateLimitLinearJitterBackoff
	client.CheckRetry = retryPolicy
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler
	standard := client.StandardClient()
	standard.Timeout = DefaultTimeout
	return standard
}

func retryPolicy(ctx context.Context, response *http.Response, err error) (bool, error) {
	if response != nil && response.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}
	return retryablehttp.DefaultRetryPolicy(ctx, response, err)
}
