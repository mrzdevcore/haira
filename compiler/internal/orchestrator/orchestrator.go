package orchestrator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/haira-lang/haira/internal/driver"
	"github.com/haira-lang/haira/internal/runtime"
)

// Orchestrator is the main deployment service.
type Orchestrator struct {
	store     *Store
	procs     *ProcessManager
	proxy     *ReverseProxy
	mux       *http.ServeMux
	dataDir   string
	indexHTML string // cached console HTML template
	uiJS     string // cached haira-ui.js bundle
}

// New creates a new orchestration service.
func New(dataDir string) (*Orchestrator, error) {
	// Ensure directory structure
	for _, dir := range []string{
		dataDir,
		filepath.Join(dataDir, "deployments"),
		filepath.Join(dataDir, "logs"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}

	store, err := NewStore(filepath.Join(dataDir, "orchestrator.db"))
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	procs := NewProcessManager(dataDir, 9100)
	proxy := NewReverseProxy()
	mux := http.NewServeMux()

	// Load UI assets from embedded runtime bundle
	uiFiles := runtime.UIFiles()
	indexHTML := string(uiFiles["ui/index.html"])
	uiJS := string(uiFiles["ui/dist/haira-ui.js"])

	o := &Orchestrator{
		store:     store,
		procs:     procs,
		proxy:     proxy,
		mux:       mux,
		dataDir:   dataDir,
		indexHTML:  indexHTML,
		uiJS:      uiJS,
	}

	// Console UI routes (must be before the catch-all proxy)
	mux.HandleFunc("/_ui/assets/haira-ui.js", o.handleConsoleJS)
	mux.HandleFunc("/_ui/", o.handleConsoleIndex)

	// API routes
	mux.HandleFunc("/_api/deploy", o.handleDeploy)
	mux.HandleFunc("/_api/deployments", o.handleListDeployments)
	mux.HandleFunc("/_api/deployments/", o.handleDeployment) // /_api/deployments/{name}[/action]

	// Everything else goes to the reverse proxy
	mux.Handle("/", proxy)

	return o, nil
}

// Serve starts the orchestration server.
func (o *Orchestrator) Serve(port int) error {
	// Restore previously running deployments
	o.restore()

	// Start health check loop
	o.procs.StartHealthLoop(o.store, 10*time.Second)

	// Graceful shutdown on signals (second signal forces immediate exit)
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Fprintf(os.Stderr, "\n[orchestrator] shutting down (ctrl+c again to force)...\n")
		go func() {
			<-sig
			fmt.Fprintf(os.Stderr, "\n[orchestrator] forced exit\n")
			os.Exit(1)
		}()
		o.shutdown()
		os.Exit(0)
	}()

	addr := fmt.Sprintf(":%d", port)
	fmt.Fprintf(os.Stderr, "[orchestrator] listening on %s\n", addr)
	fmt.Fprintf(os.Stderr, "[orchestrator] data dir: %s\n", o.dataDir)
	return http.ListenAndServe(addr, o.mux)
}

func (o *Orchestrator) restore() {
	deployments, err := o.store.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[orchestrator] restore: %v\n", err)
		return
	}

	for _, d := range deployments {
		if d.Status != "running" {
			continue
		}

		// Track highest port for allocation
		o.procs.SetNextPort(d.Port)

		proc, err := o.procs.Start(d.Name, d.BinaryPath, d.Port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[orchestrator] restore %s: %v\n", d.Name, err)
			d.Status = "crashed"
			o.store.Update(d)
			continue
		}

		d.PID = proc.Cmd.Process.Pid
		o.store.Update(d)
		o.proxy.AddRoute(d.Name, d.Port)

		fmt.Fprintf(os.Stderr, "[orchestrator] restored: %s on :%d\n", d.Name, d.Port)
	}
}

func (o *Orchestrator) shutdown() {
	// Update all running deployments to "running" so they restore on next start
	// (PIDs will be stale but that's OK — restore re-launches)
	o.procs.StopAll()
	o.store.Close()
}

// --- Console UI Handlers ---

func (o *Orchestrator) handleConsoleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/_ui/" {
		http.NotFound(w, r)
		return
	}

	// Build deployment list for metadata
	deployments, _ := o.store.List()
	type depItem struct {
		Name      string    `json:"name"`
		Status    string    `json:"status"`
		Port      int       `json:"port"`
		PID       int       `json:"pid"`
		URL       string    `json:"url"`
		Restarts  int       `json:"restarts"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	items := make([]depItem, 0, len(deployments))
	for _, d := range deployments {
		status := d.Status
		pid := d.PID
		if status == "running" && !o.procs.IsRunning(d.Name) {
			status = "crashed"
		}
		if p := o.procs.PIDByName(d.Name); p != 0 {
			pid = p
		}
		items = append(items, depItem{
			Name:      d.Name,
			Status:    status,
			Port:      d.Port,
			PID:       pid,
			URL:       fmt.Sprintf("/%s/", d.Name),
			Restarts:  d.Restarts,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		})
	}

	meta := map[string]any{
		"mode":        "orchestrator",
		"title":       "Haira Orchestrator",
		"deployments": items,
	}
	metaJSON, _ := json.Marshal(meta)
	html := strings.Replace(o.indexHTML, "{{META}}", string(metaJSON), 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (o *Orchestrator) handleConsoleJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write([]byte(o.uiJS))
}

// --- API Handlers ---

type deployRequest struct {
	Name       string `json:"name"`
	SourcePath string `json:"source_path"`
}

type deployResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Port   int    `json:"port"`
	URL    string `json:"url"`
}

func (o *Orchestrator) handleDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req deployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.SourcePath == "" {
		jsonError(w, "name and source_path are required", http.StatusBadRequest)
		return
	}

	// Validate source file exists
	if _, err := os.Stat(req.SourcePath); err != nil {
		jsonError(w, fmt.Sprintf("source file not found: %s", req.SourcePath), http.StatusBadRequest)
		return
	}

	// Check for existing deployment (redeploy)
	existing, _ := o.store.GetByName(req.Name)
	isRedeploy := existing != nil

	deployDir := filepath.Join(o.dataDir, "deployments", req.Name)
	os.MkdirAll(deployDir, 0o755)

	// Copy source file
	sourceDest := filepath.Join(deployDir, "source.haira")
	if err := copyFile(req.SourcePath, sourceDest); err != nil {
		jsonError(w, fmt.Sprintf("copy source: %v", err), http.StatusInternalServerError)
		return
	}

	// Compile
	binaryPath := filepath.Join(deployDir, "binary")
	if isRedeploy {
		binaryPath = filepath.Join(deployDir, "binary.new")
	}

	fmt.Fprintf(os.Stderr, "[orchestrator] compiling %s...\n", req.Name)
	if err := driver.Compile(req.SourcePath, binaryPath); err != nil {
		jsonError(w, fmt.Sprintf("compilation failed: %v", err), http.StatusUnprocessableEntity)
		return
	}

	if isRedeploy {
		// Hot redeploy: stop old, swap binary, start new
		o.procs.Stop(req.Name)
		o.proxy.RemoveRoute(req.Name)

		finalPath := filepath.Join(deployDir, "binary")
		os.Remove(finalPath)
		os.Rename(binaryPath, finalPath)
		binaryPath = finalPath

		existing.BinaryPath = binaryPath
		existing.SourcePath = req.SourcePath
		existing.Status = "running"
		existing.Restarts = 0
		existing.UpdatedAt = time.Now()

		proc, err := o.procs.Start(req.Name, binaryPath, existing.Port)
		if err != nil {
			existing.Status = "crashed"
			o.store.Update(existing)
			jsonError(w, fmt.Sprintf("start failed: %v", err), http.StatusInternalServerError)
			return
		}

		existing.PID = proc.Cmd.Process.Pid
		o.store.Update(existing)
		o.proxy.AddRoute(req.Name, existing.Port)

		// Wait for healthy
		if !o.procs.WaitForHealthy(req.Name, 10*time.Second) {
			fmt.Fprintf(os.Stderr, "[orchestrator] warning: %s did not pass health check\n", req.Name)
		}

		fmt.Fprintf(os.Stderr, "[orchestrator] redeployed: %s on :%d\n", req.Name, existing.Port)
		jsonResponse(w, deployResponse{
			Name:   req.Name,
			Status: "running",
			Port:   existing.Port,
			URL:    fmt.Sprintf("/%s/", req.Name),
		})
		return
	}

	// Fresh deploy
	port := o.procs.AllocatePort()

	d := &Deployment{
		ID:         fmt.Sprintf("dep_%d", time.Now().UnixNano()),
		Name:       req.Name,
		SourcePath: req.SourcePath,
		BinaryPath: binaryPath,
		Port:       port,
		Status:     "deploying",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := o.store.Create(d); err != nil {
		jsonError(w, fmt.Sprintf("store: %v", err), http.StatusInternalServerError)
		return
	}

	proc, err := o.procs.Start(req.Name, binaryPath, port)
	if err != nil {
		d.Status = "crashed"
		o.store.Update(d)
		jsonError(w, fmt.Sprintf("start failed: %v", err), http.StatusInternalServerError)
		return
	}

	d.PID = proc.Cmd.Process.Pid
	d.Status = "running"
	o.store.Update(d)
	o.proxy.AddRoute(req.Name, port)

	// Wait for healthy
	if !o.procs.WaitForHealthy(req.Name, 10*time.Second) {
		fmt.Fprintf(os.Stderr, "[orchestrator] warning: %s did not pass health check\n", req.Name)
	}

	fmt.Fprintf(os.Stderr, "[orchestrator] deployed: %s on :%d\n", req.Name, port)
	jsonResponse(w, deployResponse{
		Name:   req.Name,
		Status: "running",
		Port:   port,
		URL:    fmt.Sprintf("/%s/", req.Name),
	})
}

func (o *Orchestrator) handleListDeployments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	deployments, err := o.store.List()
	if err != nil {
		jsonError(w, fmt.Sprintf("list: %v", err), http.StatusInternalServerError)
		return
	}

	// Enrich with live status
	type entry struct {
		Name      string    `json:"name"`
		Status    string    `json:"status"`
		Port      int       `json:"port"`
		PID       int       `json:"pid"`
		URL       string    `json:"url"`
		Restarts  int       `json:"restarts"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	out := make([]entry, len(deployments))
	for i, d := range deployments {
		status := d.Status
		pid := d.PID
		if status == "running" && !o.procs.IsRunning(d.Name) {
			status = "crashed"
		}
		if p := o.procs.PIDByName(d.Name); p != 0 {
			pid = p
		}
		out[i] = entry{
			Name:      d.Name,
			Status:    status,
			Port:      d.Port,
			PID:       pid,
			URL:       fmt.Sprintf("/%s/", d.Name),
			Restarts:  d.Restarts,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		}
	}

	jsonResponse(w, out)
}

func (o *Orchestrator) handleDeployment(w http.ResponseWriter, r *http.Request) {
	// Parse: /_api/deployments/{name}[/action]
	path := strings.TrimPrefix(r.URL.Path, "/_api/deployments/")
	parts := strings.SplitN(path, "/", 2)
	name := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if name == "" {
		jsonError(w, "deployment name required", http.StatusBadRequest)
		return
	}

	d, err := o.store.GetByName(name)
	if err != nil {
		jsonError(w, fmt.Sprintf("store: %v", err), http.StatusInternalServerError)
		return
	}
	if d == nil {
		jsonError(w, fmt.Sprintf("deployment %q not found", name), http.StatusNotFound)
		return
	}

	switch {
	case action == "stop" && r.Method == http.MethodPost:
		o.procs.Stop(name)
		o.proxy.RemoveRoute(name)
		d.Status = "stopped"
		d.PID = 0
		o.store.Update(d)
		jsonResponse(w, map[string]string{"status": "stopped"})

	case action == "restart" && r.Method == http.MethodPost:
		o.procs.Stop(name)
		proc, err := o.procs.Start(name, d.BinaryPath, d.Port)
		if err != nil {
			d.Status = "crashed"
			o.store.Update(d)
			jsonError(w, fmt.Sprintf("restart failed: %v", err), http.StatusInternalServerError)
			return
		}
		d.PID = proc.Cmd.Process.Pid
		d.Status = "running"
		d.Restarts = 0
		o.store.Update(d)
		o.proxy.AddRoute(name, d.Port)

		if !o.procs.WaitForHealthy(name, 10*time.Second) {
			fmt.Fprintf(os.Stderr, "[orchestrator] warning: %s did not pass health check after restart\n", name)
		}
		jsonResponse(w, map[string]string{"status": "running"})

	case action == "logs" && r.Method == http.MethodGet:
		logPath := o.procs.LogPath(name)
		follow := r.URL.Query().Get("follow") == "true"

		if follow {
			o.streamLogs(w, r, logPath)
		} else {
			data, err := os.ReadFile(logPath)
			if err != nil {
				jsonError(w, "no logs available", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write(data)
		}

	case action == "" && r.Method == http.MethodGet:
		// Get single deployment
		status := d.Status
		if status == "running" && !o.procs.IsRunning(name) {
			status = "crashed"
		}
		jsonResponse(w, map[string]any{
			"name":       d.Name,
			"status":     status,
			"port":       d.Port,
			"pid":        d.PID,
			"url":        fmt.Sprintf("/%s/", d.Name),
			"restarts":   d.Restarts,
			"created_at": d.CreatedAt,
			"updated_at": d.UpdatedAt,
		})

	case action == "" && r.Method == http.MethodDelete:
		// Undeploy
		o.procs.Stop(name)
		o.proxy.RemoveRoute(name)

		// Remove files
		deployDir := filepath.Join(o.dataDir, "deployments", name)
		os.RemoveAll(deployDir)
		os.Remove(o.procs.LogPath(name))

		o.store.Delete(name)
		jsonResponse(w, map[string]string{"status": "removed"})

	default:
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
	}
}

func (o *Orchestrator) streamLogs(w http.ResponseWriter, r *http.Request, logPath string) {
	f, err := os.Open(logPath)
	if err != nil {
		jsonError(w, "no logs available", http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Read existing content
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			fmt.Fprintf(w, "data: %s\n\n", strings.ReplaceAll(string(buf[:n]), "\n", "\ndata: "))
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}

	// Tail new content
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, _ := f.Read(buf)
			if n > 0 {
				fmt.Fprintf(w, "data: %s\n\n", strings.ReplaceAll(string(buf[:n]), "\n", "\ndata: "))
				flusher.Flush()
			}
		}
	}
}

// --- Helpers ---

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
