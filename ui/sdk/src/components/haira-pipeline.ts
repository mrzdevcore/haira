import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import type { StepEvent, StepLogEntry } from "../core/types";
import type { HairaStep } from "./haira-step";

@customElement("haira-pipeline")
export class HairaPipeline extends LitElement {
  static styles = [
    baseStyles,
    css`
      :host {
        display: block;
        animation: fadeSlideUp 0.2s ease-out;
      }
      :host([hidden]) {
        display: none;
      }

      .pipeline {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .pipeline-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.55rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .pipeline-title {
        font-size: 0.82rem;
        font-weight: 600;
        color: var(--haira-text);
      }
      .pipeline-badge {
        font-size: 0.68rem;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
      }

      .steps {
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
        padding: 0.6rem;
      }

      /* Summary bar */
      .summary {
        display: flex;
        align-items: center;
        gap: 0.65rem;
        padding: 0.55rem 0.85rem;
        border-top: 1px solid var(--haira-border);
        background: var(--haira-bg);
        font-size: 0.76rem;
        animation: fadeIn 0.2s ease-out;
      }
      .summary-item {
        display: flex;
        align-items: center;
        gap: 0.25rem;
      }
      .summary-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
      }
      .summary-dot.done {
        background: var(--haira-success);
      }
      .summary-dot.failed {
        background: var(--haira-error);
      }
      .summary-dot.skipped {
        background: var(--haira-muted);
      }
      .summary-label {
        color: var(--haira-text-dim);
      }
      .summary-count {
        font-weight: 700;
        color: var(--haira-text);
      }
      .summary-time {
        margin-left: auto;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
      }
    `,
  ];

  @state() private _stepNames: string[] = [];
  @state() private _finalized: boolean = false;
  @state() private _totalTime: number = 0;
  @state() private _doneCt: number = 0;
  @state() private _failedCt: number = 0;
  @state() private _skippedCt: number = 0;

  private _stepStartTime: number = 0;

  /** Initialize the pipeline with step names */
  public setSteps(names: string[]): void {
    this._stepNames = names;
    this._finalized = false;
    this._totalTime = 0;
    this._doneCt = 0;
    this._failedCt = 0;
    this._skippedCt = 0;
    this._stepStartTime = Date.now();
    this.removeAttribute("hidden");
    this.requestUpdate();

    // Wait for render then set initial pending state
    this.updateComplete.then(() => {
      // Steps are rendered by Lit in the template; no imperative creation needed
    });
  }

  /** Update a step with a StepEvent */
  public updateStep(event: StepEvent): void {
    this.updateComplete.then(() => {
      const steps = this.renderRoot.querySelectorAll("haira-step");
      const step = Array.from(steps).find(
        (s) => (s as HairaStep).name === event.name
      ) as HairaStep | undefined;

      if (!step) return;

      switch (event.status) {
        case "start":
          step.setStatus("running");
          break;
        case "end":
          step.setStatus("done", event.duration_ms);
          break;
        case "failed":
          step.setStatus("failed", event.duration_ms, event.error);
          break;
        case "retry":
          step.setStatus("retrying");
          if (event.attempt != null) {
            step.addLog({
              level: "warn",
              message: `Retry attempt ${event.attempt}${event.delay_ms ? ` (delay: ${event.delay_ms}ms)` : ""}`,
            });
          }
          break;
        case "log":
          if (event.log) {
            step.addLog(event.log);
          }
          break;
      }
    });
  }

  /** Mark pipeline as complete and show summary */
  public finalize(): void {
    this._totalTime = Date.now() - this._stepStartTime;
    this._finalized = true;

    // Count statuses from step elements
    this.updateComplete.then(() => {
      const steps = this.renderRoot.querySelectorAll("haira-step");
      let done = 0;
      let failed = 0;
      let skipped = 0;

      steps.forEach((s) => {
        const stepEl = s as HairaStep & { _status?: string };
        // Read the internal status via checking classes or internal state
        // We need to access the element's status. Since _status is private,
        // we check via the reflected state in the DOM.
        const statusEl = s.renderRoot?.querySelector(".step-status-label");
        const statusText = statusEl?.textContent?.trim().toLowerCase() || "";
        if (statusText === "done") done++;
        else if (statusText === "failed") failed++;
        else if (statusText === "skipped" || statusText === "pending") skipped++;
      });

      this._doneCt = done;
      this._failedCt = failed;
      this._skippedCt = skipped;
    });
  }

  /** Reset pipeline to empty state */
  public reset(): void {
    this._stepNames = [];
    this._finalized = false;
    this._totalTime = 0;
    this._doneCt = 0;
    this._failedCt = 0;
    this._skippedCt = 0;
  }

  /** Show the pipeline */
  public show(): void {
    this.removeAttribute("hidden");
  }

  /** Hide the pipeline */
  public hide(): void {
    this.setAttribute("hidden", "");
  }

  private _formatDuration(ms: number): string {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  render() {
    if (this._stepNames.length === 0) return nothing;

    return html`
      <div class="pipeline">
        <div class="pipeline-header">
          <span class="pipeline-title">Pipeline</span>
          <span class="pipeline-badge"
            >${this._stepNames.length} step${this._stepNames.length !== 1 ? "s" : ""}</span
          >
        </div>

        <div class="steps">
          ${this._stepNames.map(
            (name) => html`<haira-step .name=${name}></haira-step>`
          )}
        </div>

        ${this._finalized
          ? html`
              <div class="summary">
                ${this._doneCt > 0
                  ? html`
                      <span class="summary-item">
                        <span class="summary-dot done"></span>
                        <span class="summary-count">${this._doneCt}</span>
                        <span class="summary-label">done</span>
                      </span>
                    `
                  : nothing}
                ${this._failedCt > 0
                  ? html`
                      <span class="summary-item">
                        <span class="summary-dot failed"></span>
                        <span class="summary-count">${this._failedCt}</span>
                        <span class="summary-label">failed</span>
                      </span>
                    `
                  : nothing}
                ${this._skippedCt > 0
                  ? html`
                      <span class="summary-item">
                        <span class="summary-dot skipped"></span>
                        <span class="summary-count">${this._skippedCt}</span>
                        <span class="summary-label">skipped</span>
                      </span>
                    `
                  : nothing}
                <span class="summary-time">
                  ${this._formatDuration(this._totalTime)}
                </span>
              </div>
            `
          : nothing}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-pipeline": HairaPipeline;
  }
}
