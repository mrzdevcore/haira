import { LitElement, html, css, nothing } from "lit";
import { customElement, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, scrollbarStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import {
  formatNumber,
  formatCost,
  formatMs,
} from "../core/utils";
import type { ObserveUsage, ObserveEvent } from "../core/types";

@customElement("haira-page-observe")
export class PageObserve extends LitElement {
  static styles = [
    baseStyles,
    scrollbarStyles,
    css`
      :host {
        display: block;
        padding: 2rem 1.5rem;
      }
      .container {
        max-width: 1200px;
        margin: 0 auto;
      }
      .page-header {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        margin-bottom: 1.5rem;
      }
      .page-desc {
        color: var(--haira-muted);
        font-size: 0.85rem;
        flex: 1;
      }
      .refresh-indicator {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        font-size: 0.72rem;
        color: var(--haira-muted);
      }
      .refresh-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: var(--haira-success);
        animation: pulse 2s ease-in-out infinite;
      }

      /* Stats grid */
      .stats-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
        gap: 0.75rem;
        margin-bottom: 2rem;
      }
      .stat-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1rem;
        animation: fadeSlideUp 0.3s ease-out both;
      }
      .stat-card.accent .stat-value {
        color: var(--haira-accent);
      }
      .stat-value {
        font-size: 1.3rem;
        font-weight: 700;
        color: var(--haira-text);
        font-family: var(--haira-mono);
      }
      .stat-label {
        font-size: 0.72rem;
        color: var(--haira-muted);
        margin-top: 0.25rem;
        text-transform: uppercase;
        letter-spacing: 0.04em;
      }

      /* Filters */
      .filters {
        display: flex;
        gap: 0.75rem;
        margin-bottom: 1.25rem;
        flex-wrap: wrap;
      }
      .filter-select,
      .filter-input {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        color: var(--haira-text);
        padding: 0.45rem 0.65rem;
        border-radius: var(--haira-radius-sm);
        font-size: 0.8rem;
        font-family: var(--haira-font);
        outline: none;
        transition: border-color 0.15s;
      }
      .filter-select:focus,
      .filter-input:focus {
        border-color: var(--haira-border-focus);
      }
      .filter-select {
        min-width: 160px;
      }
      .filter-input {
        min-width: 200px;
      }

      /* Section */
      .section-title {
        font-size: 1rem;
        font-weight: 700;
        color: var(--haira-text);
        margin-bottom: 0.75rem;
        margin-top: 1.5rem;
        display: flex;
        align-items: center;
        gap: 0.5rem;
      }

      /* Agent table */
      .agent-table {
        width: 100%;
        border-collapse: collapse;
        margin-bottom: 1.5rem;
      }
      .agent-table th {
        text-align: left;
        font-size: 0.72rem;
        font-weight: 600;
        color: var(--haira-muted);
        text-transform: uppercase;
        letter-spacing: 0.04em;
        padding: 0.5rem 0.75rem;
        border-bottom: 1px solid var(--haira-border);
      }
      .agent-table td {
        padding: 0.55rem 0.75rem;
        font-size: 0.82rem;
        border-bottom: 1px solid var(--haira-border);
        color: var(--haira-text-dim);
        font-family: var(--haira-mono);
      }
      .agent-table td:first-child {
        font-family: var(--haira-font);
        font-weight: 500;
        color: var(--haira-text);
      }

      /* Timeline */
      .timeline {
        max-height: 500px;
        overflow-y: auto;
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
      }
      .event {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        padding: 0.6rem 0.85rem;
        display: flex;
        align-items: center;
        gap: 0.6rem;
        font-size: 0.8rem;
        animation: fadeSlideUp 0.25s ease-out both;
      }
      .event-type {
        font-size: 0.6rem;
        font-weight: 700;
        padding: 0.12rem 0.4rem;
        border-radius: 3px;
        flex-shrink: 0;
        text-transform: uppercase;
        letter-spacing: 0.03em;
      }
      .event-type.llm {
        background: rgba(59, 130, 246, 0.15);
        color: var(--haira-info);
      }
      .event-type.tool {
        background: rgba(34, 197, 94, 0.15);
        color: var(--haira-success);
      }
      .event-type.fail {
        background: rgba(239, 68, 68, 0.15);
        color: var(--haira-error);
      }
      .event-agent {
        font-weight: 500;
        color: var(--haira-text);
        min-width: 80px;
      }
      .event-detail {
        flex: 1;
        color: var(--haira-text-dim);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .event-meta {
        display: flex;
        gap: 0.6rem;
        flex-shrink: 0;
        font-size: 0.72rem;
        font-family: var(--haira-mono);
        color: var(--haira-muted);
      }

      .empty {
        text-align: center;
        padding: 2rem;
        color: var(--haira-muted);
        font-size: 0.85rem;
      }
    `,
  ];

  @state() private _usage: ObserveUsage | null = null;
  @state() private _events: ObserveEvent[] = [];
  @state() private _agentNames: string[] = [];
  @state() private _agentUsages: Map<string, ObserveUsage> = new Map();
  @state() private _filterAgent = "";
  @state() private _filterSession = "";

  private _refreshTimer: ReturnType<typeof setInterval> | null = null;
  private _debounceTimer: ReturnType<typeof setTimeout> | null = null;

  connectedCallback() {
    super.connectedCallback();
    this._refresh();
    this._refreshTimer = setInterval(() => this._refresh(), 5000);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    if (this._refreshTimer) clearInterval(this._refreshTimer);
  }

  private _buildQuery(): string {
    const params = new URLSearchParams();
    if (this._filterAgent) params.set("agent", this._filterAgent);
    if (this._filterSession) params.set("session", this._filterSession);
    return params.toString() ? `?${params}` : "";
  }

  private async _refresh() {
    const q = this._buildQuery();

    const [usageRes, eventsRes, agentsRes] = await Promise.all([
      fetch(`/_observe/api/usage${q}`).catch(() => null),
      fetch(`/_observe/api/events${q}`).catch(() => null),
      fetch("/_observe/api/agents").catch(() => null),
    ]);

    if (usageRes?.ok) {
      this._usage = await usageRes.json();
    }
    if (eventsRes?.ok) {
      this._events = (await eventsRes.json()) || [];
    }
    if (agentsRes?.ok) {
      const names: string[] = (await agentsRes.json()) || [];
      this._agentNames = names;

      // Load per-agent usage
      const usages = new Map<string, ObserveUsage>();
      await Promise.all(
        names.map(async (name) => {
          try {
            const r = await fetch(
              `/_observe/api/usage?agent=${encodeURIComponent(name)}`
            );
            if (r.ok) usages.set(name, await r.json());
          } catch {
            // ignore
          }
        })
      );
      this._agentUsages = usages;
    }
  }

  private _onFilterAgent(e: Event) {
    this._filterAgent = (e.target as HTMLSelectElement).value;
    this._refresh();
  }

  private _onFilterSession(e: Event) {
    if (this._debounceTimer) clearTimeout(this._debounceTimer);
    const value = (e.target as HTMLInputElement).value;
    this._debounceTimer = setTimeout(() => {
      this._filterSession = value;
      this._refresh();
    }, 400);
  }

  private _fmtTime(ts: string): string {
    try {
      return new Date(ts).toLocaleTimeString();
    } catch {
      return ts;
    }
  }

  render() {
    const u = this._usage;
    const events = this._events.slice(0, 100);

    return html`
      <div class="container">
        <div class="page-header">
          <p class="page-desc">
            Real-time observability for agents, LLM calls, and tool executions.
          </p>
          <div class="refresh-indicator">
            <span class="refresh-dot"></span> auto-refresh 5s
          </div>
        </div>

        <!-- Stats -->
        <div class="stats-grid">
          <div class="stat-card accent" style="animation-delay:0ms">
            <div class="stat-value">${formatNumber(u?.total_tokens)}</div>
            <div class="stat-label">Total Tokens</div>
          </div>
          <div class="stat-card" style="animation-delay:30ms">
            <div class="stat-value">${formatNumber(u?.input_tokens)}</div>
            <div class="stat-label">Input Tokens</div>
          </div>
          <div class="stat-card" style="animation-delay:60ms">
            <div class="stat-value">${formatNumber(u?.output_tokens)}</div>
            <div class="stat-label">Output Tokens</div>
          </div>
          <div class="stat-card" style="animation-delay:90ms">
            <div class="stat-value">${formatNumber(u?.llm_calls)}</div>
            <div class="stat-label">LLM Calls</div>
          </div>
          <div class="stat-card" style="animation-delay:120ms">
            <div class="stat-value">${formatNumber(u?.tool_calls)}</div>
            <div class="stat-label">Tool Calls</div>
          </div>
          <div class="stat-card" style="animation-delay:150ms">
            <div class="stat-value">
              ${u ? formatMs(u.total_latency_ms) : "\u2014"}
            </div>
            <div class="stat-label">Total Latency</div>
          </div>
          <div class="stat-card accent" style="animation-delay:180ms">
            <div class="stat-value">
              ${u ? formatCost(u.estimated_cost_usd) : "\u2014"}
            </div>
            <div class="stat-label">Est. Cost</div>
          </div>
        </div>

        <!-- Filters -->
        <div class="filters">
          <select class="filter-select" @change=${this._onFilterAgent}>
            <option value="">All agents</option>
            ${this._agentNames.map(
              (name) =>
                html`<option
                  value=${name}
                  ?selected=${this._filterAgent === name}
                >
                  ${name}
                </option>`
            )}
          </select>
          <input
            class="filter-input"
            type="text"
            placeholder="Filter by session ID..."
            .value=${this._filterSession}
            @input=${this._onFilterSession}
          />
        </div>

        <!-- Agent Table -->
        ${this._agentNames.length > 0
          ? html`
              <h2 class="section-title">Agents</h2>
              <table class="agent-table">
                <thead>
                  <tr>
                    <th>Agent</th>
                    <th>Input</th>
                    <th>Output</th>
                    <th>Total</th>
                    <th>LLM</th>
                    <th>Tools</th>
                    <th>Latency</th>
                    <th>Cost</th>
                  </tr>
                </thead>
                <tbody>
                  ${this._agentNames.map((name) => {
                    const au = this._agentUsages.get(name);
                    return html`
                      <tr>
                        <td>${name}</td>
                        <td>${formatNumber(au?.input_tokens)}</td>
                        <td>${formatNumber(au?.output_tokens)}</td>
                        <td>${formatNumber(au?.total_tokens)}</td>
                        <td>${au?.llm_calls ?? "\u2014"}</td>
                        <td>${au?.tool_calls ?? "\u2014"}</td>
                        <td>${au ? formatMs(au.total_latency_ms) : "\u2014"}</td>
                        <td>
                          ${au ? formatCost(au.estimated_cost_usd) : "\u2014"}
                        </td>
                      </tr>
                    `;
                  })}
                </tbody>
              </table>
            `
          : nothing}

        <!-- Event Timeline -->
        <h2 class="section-title">
          ${unsafeHTML(iconStrings.activity)} Event Timeline
        </h2>
        ${events.length === 0
          ? html`<div class="empty">No events recorded yet.</div>`
          : html`
              <div class="timeline">
                ${events.map(
                  (ev, i) => html`
                    <div
                      class="event"
                      style="animation-delay:${Math.min(i * 30, 300)}ms"
                    >
                      ${ev.type === "generation"
                        ? html`<span class="event-type llm">LLM</span>`
                        : ev.success === false
                          ? html`<span class="event-type fail">FAIL</span>`
                          : html`<span class="event-type tool">TOOL</span>`}
                      <span class="event-agent">${ev.agent}</span>
                      <span class="event-detail">
                        ${ev.type === "generation"
                          ? `${ev.model || ""} · ${ev.input_tokens ?? 0}→${ev.output_tokens ?? 0} tokens${ev.tool_call_count ? ` · ${ev.tool_call_count} tools` : ""}`
                          : `${ev.tool_name || ""}`}
                      </span>
                      <span class="event-meta">
                        ${ev.latency_ms != null
                          ? html`${formatMs(ev.latency_ms)}`
                          : ev.duration_ms != null
                            ? html`${formatMs(ev.duration_ms)}`
                            : nothing}
                        ${ev.cost_usd
                          ? html` · ${formatCost(ev.cost_usd)}`
                          : nothing}
                        <span>${this._fmtTime(ev.timestamp)}</span>
                      </span>
                    </div>
                  `
                )}
              </div>
            `}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-page-observe": PageObserve;
  }
}
