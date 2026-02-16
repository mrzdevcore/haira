import { baseStyles, sharedKeyframes } from "../theme";
import type { StepEvent, StepStatus } from "../types";
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
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: none;
          margin-top: 0.75rem;
        }
        :host([visible]) {
          display: block;
          animation: fadeSlideUp 0.3s ease-out;
        }
        .header {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.75rem 1rem 0.5rem;
          font-size: 0.78rem;
          font-weight: 600;
          color: var(--haira-muted);
          text-transform: uppercase;
          letter-spacing: 0.04em;
        }
        .header-line {
          flex: 1;
          height: 1px;
          background: var(--haira-border);
        }
        .pipeline {
          display: flex;
          flex-direction: column;
          gap: 6px;
          padding: 0.5rem 0;
        }
        .summary {
          display: none;
          text-align: center;
          padding: 0.65rem 1rem;
          font-size: 0.8rem;
          color: var(--haira-muted);
          border-top: 1px solid var(--haira-border);
          margin-top: 0.25rem;
        }
        .summary.visible {
          display: block;
          animation: fadeIn 0.3s ease-out;
        }
        .summary .count {
          color: var(--haira-text-dim);
          font-weight: 500;
        }
        .summary .time {
          color: var(--haira-gold);
          font-family: var(--haira-mono);
        }
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
      step.style.animationDelay = `${i * 60}ms`;
      container.appendChild(step);
      this.stepElements.push(step);
    });
  }

  updateStep(event: StepEvent) {
    const idx = this.steps.indexOf(event.name);
    if (idx === -1) return;
    const step = this.stepElements[idx];
    if (!step) return;

    // Handle log events without changing step status
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
    const skippedCount = this.stepStatuses.filter(
      (s) => s === "skipped",
    ).length;
    const totalSec = (this.totalDuration / 1000).toFixed(1);

    let text = `<span class="count">${doneCount} step${doneCount !== 1 ? "s" : ""} completed`;
    if (failedCount > 0) text += `, ${failedCount} failed`;
    if (skippedCount > 0) text += `, ${skippedCount} skipped`;
    text += `</span> in <span class="time">${totalSec}s</span>`;

    summary.innerHTML = text;
    summary.classList.add("visible");
  }

  /** Called when the workflow completes — finalize any steps still running or pending. */
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
