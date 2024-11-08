package outbound

import (
	"context"
	"fmt"
	"net/http"
)

func setupRequest(ctx context.Context, url, method string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "gator")
	return req, nil
}

func Get(ctx context.Context, feedUrl string) (*http.Response, error) {
	req, err := setupRequest(ctx, feedUrl, "GET")
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return res, err
	}
	if res.StatusCode != http.StatusOK {
		return res, fmt.Errorf("status was not 200: %v", res.Status)
	}
	return res, nil
}
