import { baseCSS } from "../core";
import type { StepEvent, StepStatus } from "../core/types";
import type { HairaStep } from "./haira-step";

export class HairaPipeline extends HTMLElement {
  private steps: string[] = [];
  private stepElements: HairaStep[] = [];
  private stepStatuses: StepStatus[] = [];
  private totalDuration = 0;

  connectedCallback() {
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseCSS}
        :host { display: none; margin-top: 1.25rem; }
        :host([visible]) { display: block; animation: fadeIn 0.2s ease-out; }
        .header {
          display: flex; align-items: center; gap: 0.5rem;
          padding: 0 0.25rem 0.6rem; font-size: 0.72rem; font-weight: 600;
          color: var(--haira-muted); text-transform: uppercase; letter-spacing: 0.06em;
        }
        .header-line { flex: 1; height: 1px; background: var(--haira-border); }
        .pipeline {
          display: flex; flex-direction: column; gap: 1px;
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius); padding: 0.35rem; overflow: hidden;
        }
        .summary {
          display: none; padding: 0.75rem 1rem; font-size: 0.78rem;
          color: var(--haira-muted); border-top: 1px solid var(--haira-border);
          margin-top: 0.5rem; border-radius: var(--haira-radius-sm);
          background: var(--haira-bg-card); text-align: center;
        }
        .summary.visible { display: block; animation: fadeIn 0.3s ease-out; }
        .summary .count { color: var(--haira-text-dim); font-weight: 500; }
        .summary .time { color: var(--haira-accent); font-family: var(--haira-mono); font-weight: 600; }
        .summary .failed-count { color: var(--haira-error); }
      </style>
      <div class="header">
        <span>Pipeline</span>
        <span class="header-line"></span>
      </div>
      <div class="pipeline" id="pipeline"></div>
      <div class="summary" id="summary"></div>
    `;
  }

  setSteps(stepNames: string[]) {
    this.steps = stepNames;
    this.stepElements = [];
    this.stepStatuses = stepNames.map(() => "pending" as StepStatus);
    this.totalDuration = 0;
    const container = this.shadowRoot?.getElementById("pipeline");
    const summary = this.shadowRoot?.getElementById("summary");
    if (!container) return;
    container.innerHTML = "";
    if (summary) {
      summary.classList.remove("visible");
      summary.textContent = "";
    }

    stepNames.forEach((name, i) => {
      const step = document.createElement("haira-step") as HairaStep;
      step.setAttribute("name", name);
      step.setAttribute("index", String(i));
      container.appendChild(step);
      this.stepElements.push(step);
    });
  }

  updateStep(event: StepEvent) {
    const idx = this.steps.indexOf(event.name);
    if (idx === -1) return;
    const step = this.stepElements[idx];
    if (!step) return;

    if (event.status === "log" && event.log) {
      step.addLog(event.log.level, event.log.message);
      return;
    }

    let status: StepStatus;
    switch (event.status) {
      case "start":
        status = "running";
        break;
      case "end":
        status = "done";
        if (event.duration_ms) this.totalDuration += event.duration_ms;
        break;
      case "failed":
        status = "failed";
        break;
      case "retry":
        status = "retrying";
        break;
      default:
        return;
    }

    this.stepStatuses[idx] = status;
    step.setStatus(status, event.duration_ms, event.error);
    this.checkCompletion();
  }

  private checkCompletion() {
    const allDone = this.stepStatuses.every(
      (s) => s === "done" || s === "failed" || s === "skipped",
    );
    if (!allDone) return;
    const summary = this.shadowRoot?.getElementById("summary");
    if (!summary) return;

    const doneCount = this.stepStatuses.filter((s) => s === "done").length;
    const failedCount = this.stepStatuses.filter((s) => s === "failed").length;
    const skippedCount = this.stepStatuses.filter((s) => s === "skipped").length;
    const totalSec = (this.totalDuration / 1000).toFixed(1);

    let text = `<span class="count">${doneCount}/${this.steps.length} steps completed`;
    if (failedCount > 0)
      text += ` &middot; <span class="failed-count">${failedCount} failed</span>`;
    if (skippedCount > 0) text += ` &middot; ${skippedCount} skipped`;
    text += `</span> &middot; <span class="time">${totalSec}s total</span>`;

    summary.innerHTML = text;
    summary.classList.add("visible");
  }

  finalize() {
    for (let i = 0; i < this.stepStatuses.length; i++) {
      const s = this.stepStatuses[i];
      if (s === "running") {
        this.stepStatuses[i] = "done";
        this.stepElements[i].setStatus("done");
      } else if (s === "pending") {
        this.stepStatuses[i] = "skipped";
        this.stepElements[i].setStatus("skipped");
      }
    }
    this.checkCompletion();
  }

  reset() {
    this.totalDuration = 0;
    this.stepStatuses = this.steps.map(() => "pending" as StepStatus);
    for (const step of this.stepElements) {
      step.clearLogs();
      step.setStatus("pending");
    }
    const summary = this.shadowRoot?.getElementById("summary");
    if (summary) {
      summary.classList.remove("visible");
      summary.innerHTML = "";
    }
  }

  show() {
    this.setAttribute("visible", "");
  }

  hide() {
    this.removeAttribute("visible");
  }
}
