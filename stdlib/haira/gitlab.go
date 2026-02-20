package haira

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GitlabClient is an authenticated GitLab API client.
type GitlabClient struct {
	client  *HTTPClient
	baseURL string
	token   string
	project string
}

// GitlabMR represents a GitLab merge request.
type GitlabMR struct {
	client    *GitlabClient
	Url       string
	Iid       string
	ProjectId string
	SourceBranch string
}

// GitlabFluentBuilder builds a chain of GitLab operations.
type GitlabFluentBuilder struct {
	client     *GitlabClient
	branchName string
	files      []gitlabFileAction
	err        error
}

type gitlabFileAction struct {
	Path    string
	Content string
}

// GitlabNewClient creates a fluent GitLab builder with a token and options.
// Options: project (string or int).
func GitlabNewClient(token string, opts map[string]any) *GitlabFluentBuilder {
	baseURL := "https://gitlab.com"
	project := ""
	if opts != nil {
		if p, ok := opts["project"]; ok {
			project = Str(p)
		}
		if u, ok := opts["url"]; ok {
			baseURL = strings.TrimRight(Str(u), "/")
		}
	}

	cl := &GitlabClient{
		client: HttpClient(baseURL+"/api/v4", map[string]any{
			"headers": map[string]any{"PRIVATE-TOKEN": token},
			"retry":   2,
			"timeout": 30000,
		}),
		baseURL: baseURL,
		token:   token,
		project: project,
	}

	return &GitlabFluentBuilder{client: cl}
}

// Branch sets the branch name and creates it from the default branch.
func (b *GitlabFluentBuilder) Branch(name string) *GitlabFluentBuilder {
	if b.err != nil {
		return b
	}
	b.branchName = name

	// Get default branch
	resp, err := b.client.client.Get("/projects/" + b.client.project)
	if err != nil {
		b.err = fmt.Errorf("get project: %w", err)
		return b
	}
	projectInfo := parseJSON(resp.Body)
	defaultBranch := strVal(projectInfo, "default_branch")
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	// Create branch
	resp, err = b.client.client.Post("/projects/"+b.client.project+"/repository/branches", map[string]any{
		"branch": name,
		"ref":    defaultBranch,
	})
	if err != nil {
		b.err = fmt.Errorf("create branch %q: %w", name, err)
		return b
	}
	// 400 means branch already exists — that's OK
	if resp.StatusCode != 201 && resp.StatusCode != 400 {
		b.err = fmt.Errorf("create branch %q: HTTP %d: %s", name, resp.StatusCode, resp.Body)
	}
	return b
}

// Commit adds files to be committed. Each argument is file content;
// filenames are generated as file_1.sql, file_2.sql, etc.
// For more control, use CommitFiles().
func (b *GitlabFluentBuilder) Commit(contents ...string) *GitlabFluentBuilder {
	if b.err != nil {
		return b
	}
	for i, content := range contents {
		if content == "" {
			continue
		}
		b.files = append(b.files, gitlabFileAction{
			Path:    fmt.Sprintf("file_%d.sql", i+1),
			Content: content,
		})
	}
	return b
}

// CommitFiles adds named files to be committed.
func (b *GitlabFluentBuilder) CommitFiles(files map[string]any) *GitlabFluentBuilder {
	if b.err != nil {
		return b
	}
	for path, content := range files {
		b.files = append(b.files, gitlabFileAction{
			Path:    path,
			Content: Str(content),
		})
	}
	return b
}

// MergeRequest creates the commit and merge request, returning a GitlabMR.
func (b *GitlabFluentBuilder) MergeRequest(opts map[string]any) (*GitlabMR, error) {
	if b.err != nil {
		return nil, b.err
	}

	title := "Merge request"
	if opts != nil {
		if t, ok := opts["title"]; ok {
			title = Str(t)
		}
	}

	// Commit files via the repository commits API
	if len(b.files) > 0 {
		actions := make([]map[string]any, len(b.files))
		for i, f := range b.files {
			actions[i] = map[string]any{
				"action":    "create",
				"file_path": f.Path,
				"content":   f.Content,
			}
		}
		resp, err := b.client.client.Post("/projects/"+b.client.project+"/repository/commits", map[string]any{
			"branch":         b.branchName,
			"commit_message": title,
			"actions":        actions,
		})
		if err != nil {
			return nil, fmt.Errorf("commit files: %w", err)
		}
		if resp.StatusCode != 201 {
			return nil, fmt.Errorf("commit files: HTTP %d: %s", resp.StatusCode, resp.Body)
		}
	}

	// Create merge request
	resp, err := b.client.client.Post("/projects/"+b.client.project+"/merge_requests", map[string]any{
		"source_branch": b.branchName,
		"target_branch": "main",
		"title":         title,
	})
	if err != nil {
		return nil, fmt.Errorf("create MR: %w", err)
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("create MR: HTTP %d: %s", resp.StatusCode, resp.Body)
	}

	mrData := parseJSON(resp.Body)
	return &GitlabMR{
		client:       b.client,
		Url:          strVal(mrData, "web_url"),
		Iid:          Str(mrData["iid"]),
		ProjectId:    b.client.project,
		SourceBranch: b.branchName,
	}, nil
}

// WaitForPipeline polls the latest pipeline on the MR's source branch until it completes.
func (mr *GitlabMR) WaitForPipeline() error {
	for i := 0; i < 120; i++ { // max 60 minutes at 30s intervals
		resp, err := mr.client.client.Get(
			fmt.Sprintf("/projects/%s/merge_requests/%s/pipelines", mr.ProjectId, mr.Iid))
		if err != nil {
			return fmt.Errorf("check pipeline: %w", err)
		}
		var pipelines []map[string]any
		json.Unmarshal([]byte(resp.Body), &pipelines)
		if len(pipelines) > 0 {
			status := Str(pipelines[0]["status"])
			switch status {
			case "success":
				return nil
			case "failed", "canceled":
				return fmt.Errorf("pipeline %s: %s", Str(pipelines[0]["id"]), status)
			}
		}
		time.Sleep(30 * time.Second)
	}
	return fmt.Errorf("pipeline timed out")
}

// WaitForMerge polls the MR until it's merged or closed.
func (mr *GitlabMR) WaitForMerge() error {
	for i := 0; i < 120; i++ {
		resp, err := mr.client.client.Get(
			fmt.Sprintf("/projects/%s/merge_requests/%s", mr.ProjectId, mr.Iid))
		if err != nil {
			return fmt.Errorf("check MR: %w", err)
		}
		mrData := parseJSON(resp.Body)
		state := strVal(mrData, "state")
		switch state {
		case "merged":
			return nil
		case "closed":
			return fmt.Errorf("MR %s was closed", mr.Iid)
		}
		time.Sleep(30 * time.Second)
	}
	return fmt.Errorf("MR merge timed out")
}

// --- Direct (non-fluent) client ---

// GitlabConnect creates a direct GitLab client.
func GitlabConnect(baseURL, token string, opts map[string]any) *GitlabClient {
	project := ""
	if opts != nil {
		if p, ok := opts["project"]; ok {
			project = Str(p)
		}
	}
	return &GitlabClient{
		client: HttpClient(strings.TrimRight(baseURL, "/")+"/api/v4", map[string]any{
			"headers": map[string]any{"PRIVATE-TOKEN": token},
			"retry":   2,
			"timeout": 30000,
		}),
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		project: project,
	}
}

// CreateBranch creates a new branch.
func (gl *GitlabClient) CreateBranch(name string, opts ...map[string]any) error {
	ref := "main"
	if len(opts) > 0 && opts[0] != nil {
		if r, ok := opts[0]["ref"]; ok {
			ref = Str(r)
		}
	}
	resp, err := gl.client.Post("/projects/"+gl.project+"/repository/branches", map[string]any{
		"branch": name,
		"ref":    ref,
	})
	if err != nil {
		return fmt.Errorf("create branch: %w", err)
	}
	if resp.StatusCode != 201 && resp.StatusCode != 400 {
		return fmt.Errorf("create branch %q: HTTP %d: %s", name, resp.StatusCode, resp.Body)
	}
	return nil
}

// CommitFile commits a single file to a branch.
func (gl *GitlabClient) CommitFile(path, content, message string, opts map[string]any) error {
	branch := "main"
	if opts != nil {
		if b, ok := opts["branch"]; ok {
			branch = Str(b)
		}
	}
	resp, err := gl.client.Post("/projects/"+gl.project+"/repository/commits", map[string]any{
		"branch":         branch,
		"commit_message": message,
		"actions": []map[string]any{{
			"action":    "create",
			"file_path": path,
			"content":   content,
		}},
	})
	if err != nil {
		return fmt.Errorf("commit file: %w", err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("commit file: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

// CreateMr creates a merge request.
func (gl *GitlabClient) CreateMr(opts map[string]any) (*GitlabMR, error) {
	source := "main"
	target := "main"
	title := "Merge request"
	if opts != nil {
		if s, ok := opts["source"]; ok {
			source = Str(s)
		}
		if t, ok := opts["target"]; ok {
			target = Str(t)
		}
		if t, ok := opts["title"]; ok {
			title = Str(t)
		}
	}
	resp, err := gl.client.Post("/projects/"+gl.project+"/merge_requests", map[string]any{
		"source_branch": source,
		"target_branch": target,
		"title":         title,
	})
	if err != nil {
		return nil, fmt.Errorf("create MR: %w", err)
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("create MR: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	mrData := parseJSON(resp.Body)
	return &GitlabMR{
		client:       gl,
		Url:          strVal(mrData, "web_url"),
		Iid:          Str(mrData["iid"]),
		ProjectId:    gl.project,
		SourceBranch: source,
	}, nil
}

