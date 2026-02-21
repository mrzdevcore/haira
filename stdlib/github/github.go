package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	haira "haira-go-runtime/haira"
)

// GithubClient is an authenticated GitHub API client.
type GithubClient struct {
	client *haira.HTTPClient
	owner  string
	repo   string
}

// GithubPR represents a GitHub pull request.
type GithubPR struct {
	client *GithubClient
	Url    string
	Number string
	Branch string
}

// GithubIssue represents a GitHub issue.
type GithubIssue struct {
	Number string
	Url    string
}

// GithubFluentBuilder builds a chain of GitHub operations.
type GithubFluentBuilder struct {
	client     *GithubClient
	branchName string
	files      []githubFileAction
	err        error
}

type githubFileAction struct {
	Path    string
	Content string
}

// GithubNewClient creates a fluent GitHub builder with a token and options.
// Options: owner (string), repo (string).
func GithubNewClient(token string, opts map[string]any) *GithubFluentBuilder {
	owner, repo := "", ""
	if opts != nil {
		if o, ok := opts["owner"]; ok {
			owner = haira.Str(o)
		}
		if r, ok := opts["repo"]; ok {
			repo = haira.Str(r)
		}
	}

	cl := &GithubClient{
		client: haira.HttpClient("https://api.github.com", map[string]any{
			"headers": map[string]any{
				"Authorization": "Bearer " + token,
				"Accept":        "application/vnd.github+json",
			},
			"retry":   2,
			"timeout": 30000,
		}),
		owner: owner,
		repo:  repo,
	}

	return &GithubFluentBuilder{client: cl}
}

// Branch sets the branch name and creates it from the default branch.
func (b *GithubFluentBuilder) Branch(name string) *GithubFluentBuilder {
	if b.err != nil {
		return b
	}
	b.branchName = name

	// Get default branch SHA
	resp, err := b.client.client.Get(fmt.Sprintf("/repos/%s/%s/git/ref/heads/main", b.client.owner, b.client.repo))
	if err != nil {
		b.err = fmt.Errorf("get default branch: %w", err)
		return b
	}
	refData := haira.ParseJSON(resp.Body)
	sha := ""
	if obj, ok := refData["object"].(map[string]any); ok {
		sha = haira.Str(obj["sha"])
	}
	if sha == "" {
		b.err = fmt.Errorf("could not get default branch SHA")
		return b
	}

	// Create branch ref
	resp, err = b.client.client.Post(fmt.Sprintf("/repos/%s/%s/git/refs", b.client.owner, b.client.repo), map[string]any{
		"ref": "refs/heads/" + name,
		"sha": sha,
	})
	if err != nil {
		b.err = fmt.Errorf("create branch %q: %w", name, err)
		return b
	}
	// 422 means ref already exists -- OK
	if resp.StatusCode != 201 && resp.StatusCode != 422 {
		b.err = fmt.Errorf("create branch %q: HTTP %d: %s", name, resp.StatusCode, resp.Body)
	}
	return b
}

// Commit adds files to be committed.
func (b *GithubFluentBuilder) Commit(contents ...string) *GithubFluentBuilder {
	if b.err != nil {
		return b
	}
	for i, content := range contents {
		if content == "" {
			continue
		}
		b.files = append(b.files, githubFileAction{
			Path:    fmt.Sprintf("file_%d.sql", i+1),
			Content: content,
		})
	}
	return b
}

// CommitFiles adds named files to be committed.
func (b *GithubFluentBuilder) CommitFiles(files map[string]any) *GithubFluentBuilder {
	if b.err != nil {
		return b
	}
	for path, content := range files {
		b.files = append(b.files, githubFileAction{
			Path:    path,
			Content: haira.Str(content),
		})
	}
	return b
}

// PullRequest creates the commits and PR, returning a GithubPR.
func (b *GithubFluentBuilder) PullRequest(opts map[string]any) (*GithubPR, error) {
	if b.err != nil {
		return nil, b.err
	}

	title := "Pull request"
	if opts != nil {
		if t, ok := opts["title"]; ok {
			title = haira.Str(t)
		}
	}

	// Commit each file via the contents API
	for _, f := range b.files {
		resp, err := b.client.client.Put(
			fmt.Sprintf("/repos/%s/%s/contents/%s", b.client.owner, b.client.repo, f.Path),
			map[string]any{
				"message": title,
				"content": base64Encode(f.Content),
				"branch":  b.branchName,
			})
		if err != nil {
			return nil, fmt.Errorf("commit file %q: %w", f.Path, err)
		}
		if resp.StatusCode != 201 && resp.StatusCode != 200 {
			return nil, fmt.Errorf("commit file %q: HTTP %d: %s", f.Path, resp.StatusCode, resp.Body)
		}
	}

	// Create PR
	resp, err := b.client.client.Post(fmt.Sprintf("/repos/%s/%s/pulls", b.client.owner, b.client.repo), map[string]any{
		"title": title,
		"head":  b.branchName,
		"base":  "main",
	})
	if err != nil {
		return nil, fmt.Errorf("create PR: %w", err)
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("create PR: HTTP %d: %s", resp.StatusCode, resp.Body)
	}

	prData := haira.ParseJSON(resp.Body)
	return &GithubPR{
		client: b.client,
		Url:    haira.StrVal(prData, "html_url"),
		Number: haira.Str(prData["number"]),
		Branch: b.branchName,
	}, nil
}

// WaitForChecks polls the PR's checks until they all complete.
func (pr *GithubPR) WaitForChecks() error {
	for i := 0; i < 120; i++ {
		resp, err := pr.client.client.Get(
			fmt.Sprintf("/repos/%s/%s/commits/%s/check-runs", pr.client.owner, pr.client.repo, pr.Branch))
		if err != nil {
			return fmt.Errorf("check runs: %w", err)
		}

		var result map[string]any
		json.Unmarshal([]byte(resp.Body), &result)

		allComplete := true
		anyFailed := false
		if runs, ok := result["check_runs"].([]any); ok && len(runs) > 0 {
			for _, run := range runs {
				if r, ok := run.(map[string]any); ok {
					status := haira.Str(r["status"])
					conclusion := haira.Str(r["conclusion"])
					if status != "completed" {
						allComplete = false
					} else if conclusion == "failure" || conclusion == "cancelled" {
						anyFailed = true
					}
				}
			}
		} else {
			allComplete = false
		}

		if allComplete && anyFailed {
			return fmt.Errorf("checks failed for PR #%s", pr.Number)
		}
		if allComplete {
			return nil
		}
		time.Sleep(30 * time.Second)
	}
	return fmt.Errorf("checks timed out for PR #%s", pr.Number)
}

// Merge merges the pull request.
func (pr *GithubPR) Merge() error {
	resp, err := pr.client.client.Put(
		fmt.Sprintf("/repos/%s/%s/pulls/%s/merge", pr.client.owner, pr.client.repo, pr.Number),
		map[string]any{"merge_method": "squash"})
	if err != nil {
		return fmt.Errorf("merge PR: %w", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("merge PR: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

// --- Direct (non-fluent) client ---

// GithubConnect creates a direct GitHub client.
func GithubConnect(token string, opts map[string]any) *GithubClient {
	owner, repo := "", ""
	if opts != nil {
		if o, ok := opts["owner"]; ok {
			owner = haira.Str(o)
		}
		if r, ok := opts["repo"]; ok {
			repo = haira.Str(r)
		}
	}
	return &GithubClient{
		client: haira.HttpClient("https://api.github.com", map[string]any{
			"headers": map[string]any{
				"Authorization": "Bearer " + token,
				"Accept":        "application/vnd.github+json",
			},
			"retry":   2,
			"timeout": 30000,
		}),
		owner: owner,
		repo:  repo,
	}
}

// CreateIssue creates a new issue.
func (gh *GithubClient) CreateIssue(opts map[string]any) (*GithubIssue, error) {
	title := ""
	body := map[string]any{}
	if opts != nil {
		if t, ok := opts["title"]; ok {
			title = haira.Str(t)
			body["title"] = title
		}
		if l, ok := opts["labels"]; ok {
			body["labels"] = l
		}
		if b, ok := opts["body"]; ok {
			body["body"] = haira.Str(b)
		}
	}
	resp, err := gh.client.Post(fmt.Sprintf("/repos/%s/%s/issues", gh.owner, gh.repo), body)
	if err != nil {
		return nil, fmt.Errorf("create issue: %w", err)
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("create issue: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	issueData := haira.ParseJSON(resp.Body)
	return &GithubIssue{
		Number: haira.Str(issueData["number"]),
		Url:    haira.StrVal(issueData, "html_url"),
	}, nil
}

// AddComment adds a comment to an issue or PR.
func (gh *GithubClient) AddComment(number any, body string) error {
	resp, err := gh.client.Post(
		fmt.Sprintf("/repos/%s/%s/issues/%s/comments", gh.owner, gh.repo, haira.Str(number)),
		map[string]any{"body": body})
	if err != nil {
		return fmt.Errorf("add comment: %w", err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("add comment: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

// CreatePr creates a pull request (non-fluent).
func (gh *GithubClient) CreatePr(opts map[string]any) (*GithubPR, error) {
	title, head, base := "Pull request", "main", "main"
	if opts != nil {
		if t, ok := opts["title"]; ok {
			title = haira.Str(t)
		}
		if h, ok := opts["head"]; ok {
			head = haira.Str(h)
		}
		if b, ok := opts["base"]; ok {
			base = haira.Str(b)
		}
	}
	resp, err := gh.client.Post(fmt.Sprintf("/repos/%s/%s/pulls", gh.owner, gh.repo), map[string]any{
		"title": title,
		"head":  head,
		"base":  base,
	})
	if err != nil {
		return nil, fmt.Errorf("create PR: %w", err)
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("create PR: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	prData := haira.ParseJSON(resp.Body)
	return &GithubPR{
		client: gh,
		Url:    haira.StrVal(prData, "html_url"),
		Number: haira.Str(prData["number"]),
		Branch: head,
	}, nil
}

// base64Encode encodes a string to base64.
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
