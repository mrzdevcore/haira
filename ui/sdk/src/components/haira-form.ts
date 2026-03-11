import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state, query } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, scrollbarStyles, methodColor } from "../core/styles";
import { iconStrings } from "../core/icons";
import { submitForm, streamSSE, connectSSE } from "../services/sse-client";
import type { WorkflowMeta, RunSummary, RunDetail, StepEvent, ToolRenderEvent } from "../core/types";
import type { HairaPipeline } from "./haira-pipeline";

/**
 * <haira-form> — Form workflow UI component.
 *
 * Renders a workflow header, dynamic form fields, submit button with spinner,
 * pipeline step visualization, and result display.
 * Supports both streaming (SSE with step events) and non-streaming modes,
 * plus run resumption from URL ?run= parameter.
 */
@customElement("haira-form")
export class HairaForm extends LitElement {
  static styles = [
    baseStyles,
    scrollbarStyles,
    css`
      :host {
        display: flex;
        flex-direction: row;
        flex: 1;
        overflow: hidden;
        position: relative;
      }

      /* ---- Sidebar ---- */
      .sidebar {
        width: 240px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        border-right: 1px solid var(--haira-border);
        background: var(--haira-bg);
        overflow: hidden;
        transition: width 0.2s, opacity 0.2s;
      }
      .sidebar.collapsed {
        width: 0;
        opacity: 0;
        pointer-events: none;
      }
      .sidebar-header {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        padding: 0.55rem 0.65rem;
        border-bottom: 1px solid var(--haira-border);
        flex-shrink: 0;
      }
      .sidebar-title {
        font-size: 0.75rem;
        font-weight: 600;
        color: var(--haira-text-dim);
        flex: 1;
      }
      .sidebar-btn {
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0.25rem;
        border-radius: 4px;
        transition: all 0.15s;
      }
      .sidebar-btn:hover {
        color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }
      .sidebar-list {
        flex: 1;
        overflow-y: auto;
        padding: 0.35rem;
        display: flex;
        flex-direction: column;
        gap: 1px;
      }
      .run-item {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        padding: 0.45rem 0.5rem;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.12s;
        color: var(--haira-text-dim);
        font-size: 0.78rem;
        line-height: 1.35;
      }
      .run-item:hover {
        background: var(--haira-bg-card);
        color: var(--haira-text);
      }
      .run-item.active {
        background: var(--haira-accent-dim);
        color: var(--haira-accent);
      }
      .run-icon {
        display: flex;
        flex-shrink: 0;
        opacity: 0.5;
      }
      .run-item.active .run-icon {
        opacity: 1;
      }
      .run-info {
        flex: 1;
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 0.1rem;
      }
      .run-label {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .run-time {
        font-size: 0.65rem;
        color: var(--haira-muted);
      }
      .run-status {
        width: 7px;
        height: 7px;
        border-radius: 50%;
        flex-shrink: 0;
        background: var(--haira-muted);
      }
      .run-status.completed { background: var(--haira-success); }
      .run-status.failed { background: var(--haira-error); }
      .run-status.running { background: var(--haira-accent); animation: pulse 1.5s ease-in-out infinite; }
      .sidebar-empty {
        padding: 1rem;
        text-align: center;
        font-size: 0.75rem;
        color: var(--haira-muted);
        opacity: 0.5;
      }

      .sidebar-toggle {
        position: absolute;
        top: 0.5rem;
        left: 0.5rem;
        z-index: 10;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        color: var(--haira-muted);
        cursor: pointer;
        display: none;
        align-items: center;
        justify-content: center;
        padding: 0.35rem;
        border-radius: 6px;
        transition: all 0.15s;
      }
      .sidebar-toggle.visible {
        display: flex;
      }
      .sidebar-toggle:hover {
        color: var(--haira-accent);
        border-color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }

      /* ---- Main content ---- */
      .main-content {
        flex: 1;
        overflow-y: auto;
        min-width: 0;
        position: relative;
      }

      /* Narrow centered wrapper for the form/header/pipeline */
      .form-inner {
        max-width: 760px;
        margin: 0 auto;
        padding: 2.5rem 1.5rem 1rem;
        box-sizing: border-box;
      }

      /* Result also stays narrow */
      .result-inner {
        max-width: 760px;
        margin: 0 auto;
        padding: 0.75rem 1.5rem 3rem;
        box-sizing: border-box;
      }

      /* --- Workflow header --- */
      .workflow-header {
        margin-bottom: 2rem;
      }
      .header-top {
        display: flex;
        align-items: flex-start;
        gap: 1rem;
        margin-bottom: 1rem;
      }
      .header-avatar {
        width: 44px;
        height: 44px;
        border-radius: 10px;
        background: var(--haira-accent);
        color: #000;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 1.15rem;
        font-weight: 800;
        flex-shrink: 0;
        letter-spacing: -0.02em;
      }
      .header-text {
        flex: 1;
        min-width: 0;
      }
      .workflow-title-row {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        flex-wrap: wrap;
        margin-bottom: 0.3rem;
      }
      .workflow-title {
        font-size: 1.45rem;
        font-weight: 700;
        color: var(--haira-text);
        letter-spacing: -0.02em;
        line-height: 1.2;
      }
      .method-badge {
        display: inline-flex;
        align-items: center;
        padding: 0.15rem 0.45rem;
        border-radius: 4px;
        font-size: 0.62rem;
        font-weight: 700;
        font-family: var(--haira-mono);
        letter-spacing: 0.05em;
        text-transform: uppercase;
        color: #fff;
        line-height: 1;
        margin-top: 0.1rem;
      }
      .workflow-description {
        font-size: 0.875rem;
        color: var(--haira-text-dim);
        line-height: 1.55;
      }

      /* --- Step breadcrumb (pipeline preview) --- */
      .step-breadcrumb {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 0;
        margin-top: 1.1rem;
        padding: 0.8rem 1rem;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
      }
      .step-crumb {
        display: flex;
        align-items: center;
        gap: 0.45rem;
        font-size: 0.76rem;
        color: var(--haira-text-dim);
        font-weight: 500;
      }
      .step-crumb-num {
        width: 18px;
        height: 18px;
        border-radius: 50%;
        background: var(--haira-bg-input);
        border: 1px solid var(--haira-border);
        color: var(--haira-muted);
        font-size: 0.65rem;
        font-weight: 700;
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
      }
      .step-crumb-sep {
        color: var(--haira-border);
        font-size: 0.7rem;
        margin: 0 0.35rem;
        flex-shrink: 0;
      }

      /* --- Form card --- */
      .form-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1.75rem;
        margin-bottom: 1rem;
        transition: opacity 0.2s ease;
      }
      .form-card.disabled {
        opacity: 0.45;
        pointer-events: none;
      }
      .form-fields {
        display: flex;
        flex-direction: column;
        gap: 1.1rem;
      }
      .form-actions {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        margin-top: 1.5rem;
        padding-top: 1.25rem;
        border-top: 1px solid var(--haira-border);
      }

      /* --- Submit button --- */
      .submit-btn {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        padding: 0.65rem 2rem;
        border: none;
        border-radius: var(--haira-radius-sm);
        background: var(--haira-accent);
        color: #000;
        font-size: 0.875rem;
        font-weight: 700;
        font-family: var(--haira-font);
        cursor: pointer;
        transition: all 0.15s ease;
        letter-spacing: 0.01em;
      }
      .submit-btn:hover:not(:disabled) {
        filter: brightness(1.12);
        transform: translateY(-1px);
        box-shadow: 0 4px 12px color-mix(in srgb, var(--haira-accent) 35%, transparent);
      }
      .submit-btn:active:not(:disabled) {
        transform: translateY(0);
        box-shadow: none;
      }
      .submit-btn:disabled {
        cursor: not-allowed;
        opacity: 0.65;
      }
      .submit-btn .spinner {
        width: 14px;
        height: 14px;
        border: 2px solid rgba(0, 0, 0, 0.2);
        border-top-color: #000;
        border-radius: 50%;
        animation: spin 0.6s linear infinite;
      }

      /* --- Reset button --- */
      .reset-btn {
        display: inline-flex;
        align-items: center;
        gap: 0.35rem;
        padding: 0.55rem 1rem;
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        background: transparent;
        color: var(--haira-text-dim);
        font-size: 0.8rem;
        font-family: var(--haira-font);
        cursor: pointer;
        transition: all 0.12s ease;
      }
      .reset-btn:hover {
        background: var(--haira-bg-input);
        color: var(--haira-text);
        border-color: var(--haira-border-focus);
      }

      /* --- Banners --- */
      .error-banner {
        background: rgba(239, 68, 68, 0.07);
        border: 1px solid rgba(239, 68, 68, 0.22);
        border-left: 3px solid var(--haira-error);
        border-radius: var(--haira-radius-sm);
        padding: 0.85rem 1rem;
        margin-bottom: 1rem;
        color: var(--haira-error);
        font-size: 0.83rem;
        line-height: 1.5;
        animation: fadeSlideUp 0.2s ease-out;
      }
      .error-banner .error-title {
        font-weight: 600;
        margin-bottom: 0.2rem;
      }
      .resume-banner {
        background: rgba(59, 130, 246, 0.07);
        border: 1px solid rgba(59, 130, 246, 0.2);
        border-radius: var(--haira-radius-sm);
        padding: 0.65rem 1rem;
        margin-bottom: 1rem;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.8rem;
        color: var(--haira-info);
        animation: fadeIn 0.3s ease-out;
      }
      .resume-banner .resume-spinner {
        width: 12px;
        height: 12px;
        border: 2px solid rgba(59, 130, 246, 0.2);
        border-top-color: var(--haira-info);
        border-radius: 50%;
        animation: spin 0.6s linear infinite;
        flex-shrink: 0;
      }

      /* --- Pipeline / result sections --- */
      .pipeline-section {
        margin-bottom: 1rem;
        animation: fadeSlideUp 0.25s ease-out;
      }
      .renders-section {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        max-width: 760px;
        margin: 0 auto;
        padding: 0 1.5rem 0.75rem;
        box-sizing: border-box;
        animation: fadeSlideUp 0.25s ease-out;
      }
      .result-section {
        animation: fadeSlideUp 0.25s ease-out;
      }

      @media (max-width: 640px) {
        .form-inner {
          padding: 1.25rem 0.75rem 0.75rem;
        }
        .result-inner {
          padding: 0.5rem 0.75rem 2rem;
        }
        .renders-section {
          padding: 0 0.75rem 0.75rem;
        }
        .form-card {
          padding: 1.25rem;
        }
        .header-avatar {
          width: 38px;
          height: 38px;
          font-size: 1rem;
        }
      }
    `,
  ];

  // --- Public properties ---

  @property({ type: Object })
  meta: WorkflowMeta | null = null;

  // --- Internal state ---

  @state() private _runs: RunSummary[] = [];
  @state() private _sidebarOpen = true;
  @state() private _isRunning = false;
  @state() private _isResuming = false;
  @state() private _runId: string | null = null;
  @state() private _steps: StepEvent[] = [];
  @state() private _renders: ToolRenderEvent[] = [];
  @state() private _result: unknown = null;
  @state() private _error: string | null = null;
  @state() private _abortController: AbortController | null = null;

  // --- DOM refs ---

  @query(".form-card") private _formCard!: HTMLElement;
  @query("haira-pipeline") private _pipeline!: HairaPipeline;

  // --- Lifecycle ---

  connectedCallback() {
    super.connectedCallback();
    this._checkRunResumption();
    this._refreshRuns();
    this.addEventListener("step-confirm", this._onStepConfirm as EventListener);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this._abortController?.abort();
    this.removeEventListener("step-confirm", this._onStepConfirm as EventListener);
  }

  /** Handle step confirmation from the pipeline's confirm buttons */
  private _onStepConfirm = async (e: CustomEvent<{ confirmed: boolean }>) => {
    if (!this._runId) return;
    try {
      await fetch(`/_api/runs/${this._runId}/confirm`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ confirmed: e.detail.confirmed }),
      });
    } catch {
      // If the confirm POST fails, the step will eventually time out on the server
    }
  };

  // --- Run resumption ---

  private async _checkRunResumption() {
    const params = new URLSearchParams(window.location.search);
    const runId = params.get("run");
    if (!runId) return;

    this._isResuming = true;
    this._isRunning = true;
    this._runId = runId;

    try {
      const resp = await fetch(`/_api/runs/${encodeURIComponent(runId)}`);
      if (!resp.ok) {
        this._error = `Failed to load run ${runId}: HTTP ${resp.status}`;
        this._isResuming = false;
        this._isRunning = false;
        return;
      }

      const detail: RunDetail = await resp.json();

      // Replay existing steps — separate render events from pipeline steps
      if (detail.steps && detail.steps.length > 0) {
        const regularSteps: StepEvent[] = [];
        const renderEvents: ToolRenderEvent[] = [];
        for (const step of detail.steps) {
          if (step.status === "render" && step.render) {
            renderEvents.push({
              tool: step.name,
              component: step.render.component,
              props: { ...step.render.props, _restored: true },
            });
          } else {
            regularSteps.push(step);
          }
        }
        this._steps = regularSteps;
        this._renders = renderEvents;
      }

      // If run already finished, show result or error
      if (detail.status === "completed") {
        this._result = detail.result;
        this._isRunning = false;
        this._isResuming = false;
        this._replayRestoredPipeline(true);
        return;
      }

      if (detail.status === "failed") {
        this._error = detail.error || "Run failed";
        this._isRunning = false;
        this._isResuming = false;
        this._replayRestoredPipeline(true);
        return;
      }

      // Still running — reconnect via SSE
      this._isResuming = false;
      this._replayRestoredPipeline(false);
      await connectSSE(`/_api/runs/${encodeURIComponent(runId)}/stream`, {
        onStep: (event) => this._handleStep(event),
        onResult: (data) => {
          this._result = data;
        },
        onError: (error) => {
          this._error = error;
        },
        onDone: () => {
          this._isRunning = false;
          this._refreshRuns();
        },
      });
    } catch (err) {
      this._error = `Failed to resume run: ${err instanceof Error ? err.message : String(err)}`;
      this._isRunning = false;
      this._isResuming = false;
    }
  }

  // --- Form submission ---

  private async _handleSubmit(e: Event) {
    e.preventDefault();

    if (!this.meta || this._isRunning) return;

    // Collect field values
    const fields = this.shadowRoot?.querySelectorAll("haira-field") as
      | NodeListOf<HTMLElement & { getValue(): unknown; name: string }>
      | undefined;

    if (!fields) return;

    const body: Record<string, unknown> = {};
    let formData: FormData | undefined;
    let hasFileData = false;

    if (this.meta.hasFile) {
      formData = new FormData();
    }

    for (const field of fields) {
      const value = field.getValue();
      const name = field.name;
      if (name && value !== undefined && value !== null && value !== "") {
        if (formData && value instanceof File) {
          formData.append(name, value);
          hasFileData = true;
        } else if (formData) {
          formData.append(name, String(value));
        }
        body[name] = value;
      }
    }

    // Reset previous results
    this._result = null;
    this._error = null;
    this._steps = [];
    this._renders = [];
    this._isRunning = true;
    this._runId = null;

    const isStreaming = this.meta.steps && this.meta.steps.length > 0;
    const url = this.meta.path;

    if (isStreaming) {
      // Streaming mode — POST with SSE
      // Wait for Lit to render haira-pipeline into the DOM, then initialize it
      this._pipelineInitialized = false;
      await this.updateComplete;
      if (this._pipeline) {
        this._pipeline.setSteps(this.meta.steps!);
        this._pipelineInitialized = true;
      }

      const controller = new AbortController();
      this._abortController = controller;

      try {
        await streamSSE(
          url,
          body,
          {
            onRunId: (id) => {
              this._runId = id;
              // Update URL with run ID for resumption
              const newUrl = new URL(window.location.href);
              newUrl.searchParams.set("run", id);
              window.history.replaceState(null, "", newUrl.toString());
            },
            onStep: (event) => this._handleStep(event),
            onResult: (data) => {
              this._result = data;
            },
            onError: (error) => {
              this._error = error;
            },
            onDone: () => {
              this._isRunning = false;
              this._abortController = null;
              this._pipeline?.finalize();
              this._refreshRuns();
            },
          },
          { formData: hasFileData ? formData : undefined, signal: controller.signal },
        );
      } catch (err) {
        if (!controller.signal.aborted) {
          this._error = `Request failed: ${err instanceof Error ? err.message : String(err)}`;
          this._isRunning = false;
          this._abortController = null;
        }
      }
    } else {
      // Non-streaming mode — regular HTTP request
      try {
        const { status, data } = await submitForm(
          url,
          this.meta.method,
          body,
          this.meta.hasFile,
          hasFileData ? formData : undefined
        );

        if (status >= 200 && status < 300) {
          this._result = data;
        } else {
          const msg =
            typeof data === "object" && data && "error" in data
              ? String((data as { error: unknown }).error)
              : typeof data === "string"
                ? data
                : `HTTP ${status}`;
          this._error = msg;
        }
      } catch (err) {
        this._error = `Request failed: ${err instanceof Error ? err.message : String(err)}`;
      } finally {
        this._isRunning = false;
        this._refreshRuns();
      }
    }
  }

  // Whether the pipeline has been initialized (setSteps called) for this run.
  private _pipelineInitialized = false;

  // --- Step handling ---

  private _handleStep(event: StepEvent) {
    // Render events: collect as UI components to display below the pipeline.
    if (event.status === "render" && event.render) {
      this._renders = [
        ...this._renders,
        { tool: event.name, component: event.render.component, props: event.render.props },
      ];
      return;
    }

    this._steps = [...this._steps, event];
    // Drive the pipeline component imperatively — it has no reactive properties.
    this.updateComplete.then(() => {
      // Guard: if setSteps wasn't called yet (e.g. pipeline rendered late),
      // initialize it on the first step event.
      if (!this._pipelineInitialized && this.meta?.steps?.length) {
        this._pipelineInitialized = true;
        this._pipeline?.setSteps(this.meta.steps);
      }
      this._pipeline?.updateStep(event);
    });
  }

  // --- Pipeline restoration ---

  /**
   * Replay saved steps into the pipeline component after restoring a run.
   * Initializes the pipeline with step names and drives each step to its
   * persisted status so the visual matches the original execution.
   */
  private async _replayRestoredPipeline(isFinished: boolean) {
    if (!this.meta?.steps?.length || this._steps.length === 0) return;

    await this.updateComplete;
    if (this._pipeline) {
      this._pipeline.setSteps(this.meta.steps);
      this._pipelineInitialized = true;

      // Wait for pipeline to render its step elements
      await this._pipeline.updateComplete;

      for (const step of this._steps) {
        this._pipeline.updateStep(step);
      }
      if (isFinished) {
        this._pipeline.finalize();
      }
    }
  }

  // --- Reset ---

  private _handleReset() {
    this._abortController?.abort();
    this._abortController = null;
    this._isRunning = false;
    this._isResuming = false;
    this._runId = null;
    this._steps = [];
    this._renders = [];
    this._result = null;
    this._error = null;
    this._pipeline?.reset();
    this._pipelineInitialized = false;

    // Clean run param from URL
    const newUrl = new URL(window.location.href);
    newUrl.searchParams.delete("run");
    window.history.replaceState(null, "", newUrl.toString());
  }

  // --- Run history ---

  private async _refreshRuns() {
    if (!this.meta) return;
    try {
      const resp = await fetch(
        `/_api/runs?workflow=${encodeURIComponent(this.meta.path)}`
      );
      if (!resp.ok) return;
      const runs: RunSummary[] = await resp.json();
      this._runs = runs || [];
    } catch {
      // Silently fail
    }
  }

  private async _selectRun(runId: string) {
    if (runId === this._runId) return;

    // Reset current state
    this._abortController?.abort();
    this._abortController = null;
    this._isRunning = false;
    this._isResuming = false;
    this._steps = [];
    this._renders = [];
    this._result = null;
    this._error = null;
    this._pipeline?.reset();
    this._pipelineInitialized = false;

    this._runId = runId;

    // Update URL
    const newUrl = new URL(window.location.href);
    newUrl.searchParams.set("run", runId);
    window.history.replaceState(null, "", newUrl.toString());

    // Load the run
    try {
      const resp = await fetch(`/_api/runs/${encodeURIComponent(runId)}`);
      if (!resp.ok) {
        this._error = `Failed to load run: HTTP ${resp.status}`;
        return;
      }
      const detail: RunDetail = await resp.json();

      // Replay steps
      if (detail.steps && detail.steps.length > 0) {
        // Separate render events from regular steps
        const regularSteps: StepEvent[] = [];
        const renderEvents: ToolRenderEvent[] = [];
        for (const step of detail.steps) {
          if (step.status === "render" && step.render) {
            renderEvents.push({
              tool: step.name,
              component: step.render.component,
              props: { ...step.render.props, _restored: true },
            });
          } else {
            regularSteps.push(step);
          }
        }
        this._steps = regularSteps;
        this._renders = renderEvents;
      }

      if (detail.status === "completed") {
        this._result = detail.result;
        this._replayRestoredPipeline(true);
      } else if (detail.status === "failed") {
        this._error = detail.error || "Run failed";
        this._replayRestoredPipeline(true);
      } else {
        // Still running — reconnect via SSE
        this._isRunning = true;
        this._replayRestoredPipeline(false);
        await connectSSE(`/_api/runs/${encodeURIComponent(runId)}/stream`, {
          onStep: (event) => this._handleStep(event),
          onResult: (data) => { this._result = data; },
          onError: (error) => { this._error = error; },
          onDone: () => {
            this._isRunning = false;
            this._refreshRuns();
          },
        });
      }
    } catch (err) {
      this._error = `Failed to load run: ${err instanceof Error ? err.message : String(err)}`;
    }
  }

  private _startNewRun() {
    this._handleReset();
    this._refreshRuns();
  }

  private _formatRunLabel(runId: string): string {
    // run_20260303_124131_005 → "Run #5 · 12:41"
    const parts = runId.split("_");
    if (parts.length >= 4) {
      const num = parseInt(parts[parts.length - 1], 10);
      const timeStr = parts[parts.length - 2]; // e.g. "124131"
      if (timeStr.length >= 4) {
        const hh = timeStr.slice(0, 2);
        const mm = timeStr.slice(2, 4);
        return `Run #${num} · ${hh}:${mm}`;
      }
      return `Run #${num}`;
    }
    return runId;
  }

  private _formatRelativeTime(dateStr: string): string {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    if (diffMins < 1) return "just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    const diffDays = Math.floor(diffHours / 24);
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  }

  // --- Render ---

  render() {
    if (!this.meta) return nothing;

    const meta = this.meta;
    const isStreaming = meta.steps && meta.steps.length > 0;
    const hasResult = this._result !== null;
    const hasError = this._error !== null;
    const showPipeline =
      isStreaming && (this._steps.length > 0 || this._isRunning);
    const showReset = hasResult || hasError || this._steps.length > 0 || this._renders.length > 0;

    return html`
      ${this._renderSidebar()}

      <div class="main-content">
        <!-- Sidebar open toggle (when collapsed) -->
        <button
          class="sidebar-toggle ${this._sidebarOpen ? "" : "visible"}"
          title="Show runs"
          @click=${() => (this._sidebarOpen = true)}
        >
          ${unsafeHTML(iconStrings.sidebar)}
        </button>

        <div class="form-inner">
          ${this._renderHeader(meta)}
          ${this._isResuming ? this._renderResumeBanner() : nothing}
          ${hasError ? this._renderError() : nothing}

          <form
            @submit=${this._handleSubmit}
            class="form-card ${this._isRunning ? "disabled" : ""}"
          >
            <div class="form-fields">
              ${(meta.params || []).map(
                (param) => html`
                  <haira-field
                    .name=${param.Name}
                    .type=${param.Type}
                    .disabled=${this._isRunning}
                  ></haira-field>
                `
              )}
            </div>
            <div class="form-actions">
              <button
                type="submit"
                class="submit-btn"
                ?disabled=${this._isRunning}
              >
                ${this._isRunning
                  ? html`<span class="spinner"></span> Running...`
                  : isStreaming ? "Run Pipeline" : "Submit"}
              </button>
              ${showReset
                ? html`
                    <button
                      type="button"
                      class="reset-btn"
                      @click=${this._handleReset}
                    >
                      Reset
                    </button>
                  `
                : nothing}
            </div>
          </form>

          ${showPipeline
            ? html`
                <div class="pipeline-section">
                  <haira-pipeline></haira-pipeline>
                </div>
              `
            : nothing}
        </div>

        ${this._renders.length > 0
          ? html`
              <div class="renders-section">
                ${this._renders.map(
                  (ev) => html`
                    <haira-ui-renderer .event=${ev}></haira-ui-renderer>
                  `
                )}
              </div>
            `
          : nothing}

        ${hasResult
          ? html`
              <div class="result-inner result-section">
                <haira-result .data=${this._result}></haira-result>
              </div>
            `
          : nothing}
      </div>
    `;
  }

  // --- Sub-renders ---

  private _renderHeader(meta: WorkflowMeta) {
    const color = methodColor(meta.method);
    const avatarChar = meta.avatar
      ? meta.avatar
      : (meta.title || meta.name || "?")[0].toUpperCase();
    const steps = meta.steps || [];

    return html`
      <div class="workflow-header">
        <div class="header-top">
          <div class="header-avatar">${avatarChar}</div>
          <div class="header-text">
            <div class="workflow-title-row">
              <span class="workflow-title">${meta.title || meta.name}</span>
              <span class="method-badge" style="background:${color}"
                >${meta.method}</span
              >
            </div>
            ${meta.description
              ? html`<div class="workflow-description">${meta.description}</div>`
              : nothing}
          </div>
        </div>
        ${steps.length > 0
          ? html`
              <div class="step-breadcrumb">
                ${steps.map(
                  (s, i) => html`
                    <div class="step-crumb">
                      <span class="step-crumb-num">${i + 1}</span>
                      <span>${s}</span>
                    </div>
                    ${i < steps.length - 1
                      ? html`<span class="step-crumb-sep">›</span>`
                      : nothing}
                  `
                )}
              </div>
            `
          : nothing}
      </div>
    `;
  }

  private _renderResumeBanner() {
    return html`
      <div class="resume-banner">
        <span class="resume-spinner"></span>
        Reconnecting to run ${this._runId ? this._runId : ""}...
      </div>
    `;
  }

  private _renderError() {
    return html`
      <div class="error-banner">
        <div class="error-title">Error</div>
        <div>${this._error}</div>
      </div>
    `;
  }

  // --- Sidebar ---

  private _renderSidebar() {
    return html`
      <div class="sidebar ${this._sidebarOpen ? "" : "collapsed"}">
        <div class="sidebar-header">
          <span class="sidebar-title">Runs</span>
          <button
            class="sidebar-btn"
            title="New run"
            @click=${this._startNewRun}
          >
            ${unsafeHTML(iconStrings.plus)}
          </button>
          <button
            class="sidebar-btn"
            title="Close sidebar"
            @click=${() => (this._sidebarOpen = false)}
          >
            ${unsafeHTML(iconStrings.chevronLeft)}
          </button>
        </div>
        <div class="sidebar-list">
          ${this._runs.length === 0
            ? html`<div class="sidebar-empty">No runs yet</div>`
            : this._runs.map(
                (run) => html`
                  <div
                    class="run-item ${run.id === this._runId ? "active" : ""}"
                    @click=${() => this._selectRun(run.id)}
                  >
                    <span class="run-icon"
                      >${unsafeHTML(iconStrings.activity)}</span
                    >
                    <span class="run-info">
                      <span class="run-label">${this._formatRunLabel(run.id)}</span>
                      <span class="run-time">${this._formatRelativeTime(run.started_at)}</span>
                    </span>
                    <span class="run-status ${run.status}"></span>
                  </div>
                `
              )}
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-form": HairaForm;
  }
}
