package backlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	SpaceURL   string
	APIKey     string
	HTTPClient *http.Client
}

type Issue struct {
	IssueKey string
	Summary  string
	Status   string
	URL      string
}

func (c Client) GetIssue(ctx context.Context, issueKey string) (Issue, error) {
	var response issueResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/v2/issues/"+url.PathEscape(issueKey), nil, &response); err != nil {
		return Issue{}, err
	}

	return Issue{
		IssueKey: response.IssueKey,
		Summary:  response.Summary,
		Status:   response.Status.Name,
		URL:      c.issueURL(response.IssueKey),
	}, nil
}

func (c Client) UpdateIssueStatus(ctx context.Context, issueKey string, statusID int) error {
	values := url.Values{}
	values.Set("statusId", fmt.Sprintf("%d", statusID))
	return c.doJSON(ctx, http.MethodPatch, "/api/v2/issues/"+url.PathEscape(issueKey), values, nil)
}

func (c Client) doJSON(ctx context.Context, method string, path string, form url.Values, target interface{}) error {
	endpoint, err := c.endpoint(path)
	if err != nil {
		return err
	}

	query := endpoint.Query()
	query.Set("apiKey", c.APIKey)
	endpoint.RawQuery = query.Encode()

	var body *strings.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return err
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Backlog API %s %s failed: %s", method, path, res.Status)
	}
	if target == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(target)
}

func (c Client) endpoint(path string) (*url.URL, error) {
	if c.SpaceURL == "" {
		return nil, fmt.Errorf("Backlog space URL is empty")
	}
	base, err := url.Parse(strings.TrimRight(c.SpaceURL, "/"))
	if err != nil {
		return nil, err
	}
	base.Path = strings.TrimRight(base.Path, "/") + path
	return base, nil
}

func (c Client) issueURL(issueKey string) string {
	return strings.TrimRight(c.SpaceURL, "/") + "/view/" + url.PathEscape(issueKey)
}

type issueResponse struct {
	IssueKey string `json:"issueKey"`
	Summary  string `json:"summary"`
	Status   struct {
		Name string `json:"name"`
	} `json:"status"`
}
