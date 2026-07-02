package backlog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetIssueFetchesIssueAndBuildsViewURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/v2/issues/COMMUNITY-101" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("apiKey") != "secret" {
			t.Fatalf("unexpected apiKey: %s", r.URL.Query().Get("apiKey"))
		}

		writeIssueResponse(t, w, issueResponse{
			IssueKey: "COMMUNITY-101",
			Summary:  "ログイン画面を実装",
			Status: struct {
				Name string `json:"name"`
			}{Name: "対応中"},
		})
	}))
	defer server.Close()

	client := Client{SpaceURL: server.URL + "/", APIKey: "secret", HTTPClient: server.Client()}
	issue, err := client.GetIssue(context.Background(), "COMMUNITY-101")
	if err != nil {
		t.Fatal(err)
	}

	if issue.IssueKey != "COMMUNITY-101" {
		t.Fatalf("unexpected issue key: %s", issue.IssueKey)
	}
	if issue.Summary != "ログイン画面を実装" {
		t.Fatalf("unexpected summary: %s", issue.Summary)
	}
	if issue.Status != "対応中" {
		t.Fatalf("unexpected status: %s", issue.Status)
	}
	if issue.URL != server.URL+"/view/COMMUNITY-101" {
		t.Fatalf("unexpected issue URL: %s", issue.URL)
	}
}

func TestUpdateIssueStatusSendsFormBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Fatalf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("statusId") != "5" {
			t.Fatalf("unexpected statusId: %s", r.Form.Get("statusId"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()}
	if err := client.UpdateIssueStatus(context.Background(), "COMMUNITY-101", 5); err != nil {
		t.Fatal(err)
	}
}

func TestGetIssueReturnsErrorForNonSuccessResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client := Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()}
	_, err := client.GetIssue(context.Background(), "COMMUNITY-101")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "502 Bad Gateway") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func writeIssueResponse(t *testing.T, w http.ResponseWriter, response issueResponse) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		t.Fatal(err)
	}
}
