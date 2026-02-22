import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state, query } from "lit/decorators.js";
import { baseStyles, methodColor } from "../core/styles";
import { submitForm, streamSSE, connectSSE } from "../services/sse-client";
import type { WorkflowMeta, RunDetail, StepEvent } from "../core/types";

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
    css`
      :host {
        display: block;
        width: 100%;
        max-width: 720px;
        margin: 0 auto;
        padding: 2rem 1.25rem;
      }

      /* --- Workflow header --- */
      .workflow-header {
        margin-bottom: 1.75rem;
      }
      .workflow-title-row {
        display: flex;
        align-items: center;
        gap: 0.6rem;
        margin-bottom: 0.35rem;
      }
      .workflow-title {
        font-size: 1.35rem;
        font-weight: 700;
        color: var(--haira-text);
        letter-spacing: -0.01em;
      }
      .method-badge {
        display: inline-flex;
        align-items: center;
        padding: 0.15rem 0.5rem;
        border-radius: 4px;
        font-size: 0.65rem;
        font-weight: 700;
        font-family: var(--haira-mono);
        letter-spacing: 0.04em;
        text-transform: uppercase;
        color: #fff;
        line-height: 1;
      }
      .workflow-description {
        font-size: 0.88rem;
        color: var(--haira-text-dim);
        line-height: 1.5;
        margin-top: 0.3rem;
      }
      .workflow-path {
        font-size: 0.75rem;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
        margin-top: 0.35rem;
      }

      /* --- Form card --- */
      .form-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1.5rem;
        margin-bottom: 1.25rem;
      }
      .form-card.disabled {
        opacity: 0.55;
        pointer-events: none;
      }
      .form-fields {
        display: flex;
        flex-direction: column;
        gap: 1rem;
      }

      /* --- Submit button --- */
      .submit-row {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        margin-top: 1.25rem;
      }
      .submit-btn {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        padding: 0.6rem 1.5rem;
        border: none;
        border-radius: var(--haira-radius-sm);
        background: var(--haira-accent);
        color: #000;
        font-size: 0.85rem;
        font-weight: 600;
        font-family: var(--haira-font);
        cursor: pointer;
        transition: all 0.15s ease;
        min-width: 120px;
      }
      .submit-btn:hover:not(:disabled) {
        filter: brightness(1.1);
        transform: translateY(-1px);
      }
      .submit-btn:active:not(:disabled) {
        transform: translateY(0);
      }
      .submit-btn:disabled {
        cursor: not-allowed;
        opacity: 0.7;
      }
      .submit-btn .spinner {
        width: 14px;
        height: 14px;
        border: 2px solid rgba(0, 0, 0, 0.2);
        border-top-color: #000;
        border-radius: 50%;
        animation: spin 0.6s linear infinite;
      }

      /* --- Error banner --- */
      .error-banner {
        background: rgba(239, 68, 68, 0.08);
        border: 1px solid rgba(239, 68, 68, 0.25);
        border-radius: var(--haira-radius-sm);
        padding: 0.75rem 1rem;
        margin-bottom: 1rem;
        color: var(--haira-error);
        font-size: 0.82rem;
        line-height: 1.45;
        animation: fadeSlideUp 0.2s ease-out;
      }
      .error-banner .error-title {
        font-weight: 600;
        margin-bottom: 0.2rem;
      }

      /* --- Pipeline section --- */
      .pipeline-section {
        margin-bottom: 1.25rem;
        animation: fadeSlideUp 0.25s ease-out;
      }

      /* --- Result section --- */
      .result-section {
        animation: fadeSlideUp 0.25s ease-out;
      }

      /* --- Resumption banner --- */
      .resume-banner {
        background: rgba(59, 130, 246, 0.08);
        border: 1px solid rgba(59, 130, 246, 0.2);
        border-radius: var(--haira-radius-sm);
        padding: 0.6rem 1rem;
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

      /* --- Reset button --- */
      .reset-btn {
        display: inline-flex;
        align-items: center;
        gap: 0.35rem;
        padding: 0.4rem 0.85rem;
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        background: transparent;
        color: var(--haira-text-dim);
        font-size: 0.78rem;
        font-family: var(--haira-font);
        cursor: pointer;
        transition: all 0.12s ease;
      }
      .reset-btn:hover {
        background: var(--haira-bg-card);
        color: var(--haira-text);
        border-color: var(--haira-border-focus);
      }

      @media (max-width: 640px) {
        :host {
          padding: 1.25rem 0.75rem;
        }
        .form-card {
          padding: 1rem;
        }
      }
    `,
  ];

  // --- Public properties ---

  @property({ type: Object })
  meta: WorkflowMeta | null = null;

  // --- Internal state ---

  @state() private _isRunning = false;
  @state() private _isResuming = false;
  @state() private _runId: string | null = null;
  @state() private _steps: StepEvent[] = [];
  @state() private _result: unknown = null;
  @state() private _error: string | null = null;
  @state() private _abortController: AbortController | null = null;

  // --- DOM refs ---

  @query(".form-card") private _formCard!: HTMLElement;

  // --- Lifecycle ---

  connectedCallback() {
    super.connectedCallback();
    this._checkRunResumption();
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this._abortController?.abort();
  }

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

      // Replay existing steps
      if (detail.steps && detail.steps.length > 0) {
        this._steps = [...detail.steps];
      }

      // If run already finished, show result or error
      if (detail.status === "completed") {
        this._result = detail.result;
        this._isRunning = false;
        this._isResuming = false;
        return;
      }

      if (detail.status === "failed") {
        this._error = detail.error || "Run failed";
        this._isRunning = false;
        this._isResuming = false;
        return;
      }

      // Still running — reconnect via SSE
      this._isResuming = false;
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
    this._isRunning = true;
    this._runId = null;

    const isStreaming = this.meta.steps && this.meta.steps.length > 0;
    const url = this.meta.path;

    if (isStreaming) {
      // Streaming mode — POST with SSE
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
            },
          },
          hasFileData ? formData : undefined,
          controller.signal
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
      }
    }
  }

  // --- Step handling ---

  private _handleStep(event: StepEvent) {
    // Accumulate step events. For start/end/failed/retry of the same step name,
    // we keep all events so the pipeline component can render the latest status.
    this._steps = [...this._steps, event];
  }

  // --- Reset ---

  private _handleReset() {
    this._abortController?.abort();
    this._abortController = null;
    this._isRunning = false;
    this._isResuming = false;
    this._runId = null;
    this._steps = [];
    this._result = null;
    this._error = null;

    // Clean run param from URL
    const newUrl = new URL(window.location.href);
    newUrl.searchParams.delete("run");
    window.history.replaceState(null, "", newUrl.toString());
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
    const showReset = hasResult || hasError || this._steps.length > 0;

    return html`
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
        <div class="submit-row">
          <button
            type="submit"
            class="submit-btn"
            ?disabled=${this._isRunning}
          >
            ${this._isRunning
              ? html`<span class="spinner"></span> Running...`
              : "Submit"}
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
              <haira-pipeline
                .steps=${meta.steps}
                .events=${this._steps}
              ></haira-pipeline>
            </div>
          `
        : nothing}
      ${hasResult
        ? html`
            <div class="result-section">
              <haira-result .data=${this._result}></haira-result>
            </div>
          `
        : nothing}
    `;
  }

  // --- Sub-renders ---

  private _renderHeader(meta: WorkflowMeta) {
    const color = methodColor(meta.method);
    return html`
      <div class="workflow-header">
        <div class="workflow-title-row">
          <span class="workflow-title"
            >${meta.title || meta.name}</span
          >
          <span
            class="method-badge"
            style="background:${color}"
            >${meta.method}</span
          >
        </div>
        ${meta.description
          ? html`<div class="workflow-description">${meta.description}</div>`
          : nothing}
        <div class="workflow-path">${meta.path}</div>
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
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-form": HairaForm;
  }
}
