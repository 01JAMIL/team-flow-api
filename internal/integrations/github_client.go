package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var errRepositoryNotFound = errors.New("repository not found")

type gitHubClient struct {
	httpClient *http.Client
}

func newGitHubClient() *gitHubClient {
	return &gitHubClient{
		httpClient: &http.Client{},
	}
}

func (c *gitHubClient) RepositoryExists(ctx context.Context, owner, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, name), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errRepositoryNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var repo gitHubRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return err
	}

	return nil
}
