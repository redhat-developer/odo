package git

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
)

var mockHeaders = map[string]string{
	"X-GitHub-Request-Id":   "DD0E:6011:12F21A8:1926790:5A2064E2",
	"X-RateLimit-Limit":     "60",
	"X-RateLimit-Remaining": "59",
	"X-RateLimit-Reset":     "1512076018",
}

func TestWebhookWithFakeClient(t *testing.T) {

	repo, err := NewRepository("https://fake.com/foo/bar.git", "token")
	if err != nil {
		t.Fatal(err)
	}

	listenerURL := "http://example.com/webhook"
	ids, err := repo.ListWebhooks(listenerURL)
	if err != nil {
		t.Fatal(err)
	}

	// start with no webhooks
	if len(ids) > 0 {
		t.Fatal(err)
	}

	// create a webhook
	id, err := repo.CreateWebhook(listenerURL, "secret")
	if len(ids) > 0 {
		t.Fatal(err)
	}

	// verify and remember our ID
	if id == "" {
		t.Fatal(err)
	}

	// list again
	ids, err = repo.ListWebhooks(listenerURL)
	if err != nil {
		t.Fatal(err)
	}

	// verify ID from list
	if diff := cmp.Diff(ids, []string{id}); diff != "" {
		t.Fatalf("created id mismatch got\n%s", diff)
	}

	// delete webhook
	deleted, err := repo.DeleteWebhooks(ids)
	if err != nil {
		t.Fatal(err)
	}

	// verify deleted IDs
	if diff := cmp.Diff(ids, deleted); diff != "" {
		t.Fatalf("deleted ids mismatch got\n%s", diff)
	}

	ids, err = repo.ListWebhooks(listenerURL)
	if err != nil {
		t.Fatal(err)
	}

	// verify no webhooks
	if len(ids) > 0 {
		t.Fatal(err)
	}
}

func TestListWebHooks(t *testing.T) {

	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/foo/bar/hooks").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/hooks.json")

	repo, err := NewRepository("https://github.com/foo/bar.git", "token")
	if err != nil {
		t.Fatal(err)
	}

	ids, err := repo.ListWebhooks("http://example.com/webhook")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(ids, []string{"1"}); diff != "" {
		t.Errorf("driver errMsg mismatch got\n%s", diff)
	}
}

func TestDeleteWebHooks(t *testing.T) {

	defer gock.Off()

	gock.New("https://api.github.com").
		Delete("/repos/foo/bar/hooks/1").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	repo, err := NewRepository("https://github.com/foo/bar.git", "token")
	if err != nil {
		t.Fatal(err)
	}

	deleted, err := repo.DeleteWebhooks([]string{"1"})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff([]string{"1"}, deleted); diff != "" {
		t.Errorf("deleted mismatch got\n%s", diff)
	}
}

func TestCreateWebHook(t *testing.T) {

	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/repos/foo/bar/hooks").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/hook.json")

	repo, err := NewRepository("https://github.com/foo/bar.git", "token")
	if err != nil {
		t.Fatal(err)
	}

	created, err := repo.CreateWebhook("http://example.com/webhook", "mysecret")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff("1", created); diff != "" {
		t.Errorf("deleted mismatch got\n%s", diff)
	}
}

func TestGetDriverName(t *testing.T) {

	tests := []struct {
		url          string
		driver       string
		driverErrMsg string
		repoName     string
		repoErrMsg   string
	}{
		{
			"http://github.org",
			"github",
			"",
			"",
			"failed to get Git repo: ",
		},
		{
			"http://github.com/",
			"github",
			"",
			"",
			"failed to get Git repo: /",
		},
		{
			"http://github.com/foo/bar",
			"github",
			"",
			"foo/bar",
			"",
		},
		{
			"https://githuB.com/foo/bar.git",
			"github",
			"",
			"foo/bar",
			"",
		},
		{
			"http://gitlab.com/foo/bar.git2",
			"gitlab",
			"",
			"",
			"failed to get Git repo: /foo/bar.git2",
		},
		{
			"http://gitlab/foo/bar/",
			"",
			"unknown Git server: gitlab",
			"foo/bar",
			"",
		},
		{
			"https://gitlab.a.b/foo/bar/bar",
			"",
			"unknown Git server: gitlab.a.b",
			"",
			"failed to get Git repo: /foo/bar/bar",
		},
		{
			"https://gitlab.org2/f.b/bar.git",
			"",
			"unknown Git server: gitlab.org2",
			"",
			"failed to get Git repo: /f.b/bar.git",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Error(err)
			} else {
				gotDriver, err := getDriverName(u)
				driverErrMsg := ""
				if err != nil {
					driverErrMsg = err.Error()
				}
				if diff := cmp.Diff(tt.driverErrMsg, driverErrMsg); diff != "" {
					t.Errorf("driver errMsg mismatch got\n%s", diff)
				}
				if diff := cmp.Diff(tt.driver, gotDriver); diff != "" {
					t.Errorf("driver mismatch got\n%s", diff)
				}

				repoName, err := getRepoName(u)
				repoErrMsg := ""
				if err != nil {
					repoErrMsg = err.Error()
				}
				if diff := cmp.Diff(tt.repoErrMsg, repoErrMsg); diff != "" {
					t.Errorf("driver errMsg mismatch got\n%s", diff)
				}
				if diff := cmp.Diff(tt.repoName, repoName); diff != "" {
					t.Errorf("driver mismatch got\n%s", diff)
				}
			}
		})
	}
}
