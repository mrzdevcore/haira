import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, scrollbarStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import { formatNumber, formatCost, formatMs } from "../core/utils";
import type { WorkflowMeta, ObserveUsage } from "../core/types";

interface AgentInfo {
  name: string;
  usage: ObserveUsage | null;
}

@customElement("haira-page-agents")
export class PageAgents extends LitElement {
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
      .page-desc {
        color: var(--haira-muted);
        font-size: 0.85rem;
        margin-bottom: 1.5rem;
      }

      .agents-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
        gap: 0.75rem;
      }
      .agent-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1.25rem;
        animation: fadeSlideUp 0.3s ease-out both;
        transition: border-color 0.15s;
      }
      .agent-card:hover {
        border-color: rgba(232, 163, 23, 0.3);
      }
      .agent-header {
        display: flex;
        align-items: center;
        gap: 0.6rem;
        margin-bottom: 0.85rem;
      }
      .agent-icon {
        width: 36px;
        height: 36px;
        border-radius: 8px;
        background: var(--haira-accent-dim);
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--haira-accent);
        flex-shrink: 0;
      }
      .agent-name {
        font-weight: 600;
        font-size: 0.95rem;
        color: var(--haira-text);
      }
      .agent-badge {
        font-size: 0.62rem;
        font-weight: 600;
        padding: 0.1rem 0.4rem;
        border-radius: 10px;
        background: rgba(34, 197, 94, 0.1);
        color: var(--haira-success);
        margin-left: auto;
      }

      .stats-row {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 0.5rem;
        margin-bottom: 0.75rem;
      }
      .stat {
        text-align: center;
        padding: 0.5rem;
        background: var(--haira-bg);
        border-radius: var(--haira-radius-sm);
      }
      .stat-value {
        font-size: 0.9rem;
        font-weight: 700;
        color: var(--haira-text);
        font-family: var(--haira-mono);
      }
      .stat-label {
        font-size: 0.65rem;
        color: var(--haira-muted);
        margin-top: 0.15rem;
        text-transform: uppercase;
        letter-spacing: 0.03em;
      }

      .agent-meta {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
        font-size: 0.72rem;
        color: var(--haira-text-dim);
      }
      .meta-chip {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        padding: 0.2rem 0.5rem;
        background: var(--haira-bg);
        border-radius: 4px;
      }
      .meta-chip .label {
        color: var(--haira-muted);
      }

      .empty {
        text-align: center;
        padding: 3rem;
        color: var(--haira-muted);
        font-size: 0.85rem;
      }
      .empty-icon {
        opacity: 0.3;
        margin-bottom: 0.75rem;
      }
    `,
  ];

  @property({ type: Object }) meta: WorkflowMeta | null = null;
  @state() private _agents: AgentInfo[] = [];
  @state() private _loading = true;

  connectedCallback() {
    super.connectedCallback();
    this._loadAgents();
  }

  private async _loadAgents() {
    try {
      const resp = await fetch("/_observe/api/agents");
      if (!resp.ok) {
        this._loading = false;
        return;
      }
      const names: string[] = await resp.json();
      if (!names || names.length === 0) {
        this._loading = false;
        return;
      }

      const agents: AgentInfo[] = await Promise.all(
        names.map(async (name) => {
          try {
            const r = await fetch(
              `/_observe/api/usage?agent=${encodeURIComponent(name)}`
            );
            if (r.ok) {
              return { name, usage: await r.json() };
            }
          } catch {
            // ignore
          }
          return { name, usage: null };
        })
      );

      this._agents = agents;
    } catch {
      // silently fail
    }
    this._loading = false;
  }

  render() {
    if (this._loading) {
      return html`<div class="container">
        <p class="page-desc">Loading agents...</p>
      </div>`;
    }

    return html`
      <div class="container">
        <p class="page-desc">
          All registered agents and their usage statistics.
        </p>

        ${this._agents.length === 0
          ? html`
              <div class="empty">
                <div class="empty-icon">
                  ${unsafeHTML(iconStrings.agent)}
                </div>
                <div>No agents have been active yet.</div>
                <div style="margin-top:0.25rem;font-size:0.78rem">
                  Agent activity will appear here once workflows run.
                </div>
              </div>
            `
          : html`
              <div class="agents-grid">
                ${this._agents.map(
                  (agent, i) => html`
                    <div
                      class="agent-card"
                      style="animation-delay:${i * 50}ms"
                    >
                      <div class="agent-header">
                        <div class="agent-icon">
                          ${unsafeHTML(iconStrings.agent)}
                        </div>
                        <span class="agent-name">${agent.name}</span>
                        <span class="agent-badge">Active</span>
                      </div>

                      ${agent.usage
                        ? html`
                            <div class="stats-row">
                              <div class="stat">
                                <div class="stat-value">
                                  ${formatNumber(agent.usage.total_tokens)}
                                </div>
                                <div class="stat-label">Tokens</div>
                              </div>
                              <div class="stat">
                                <div class="stat-value">
                                  ${agent.usage.llm_calls}
                                </div>
                                <div class="stat-label">LLM Calls</div>
                              </div>
                              <div class="stat">
                                <div class="stat-value">
                                  ${agent.usage.tool_calls}
                                </div>
                                <div class="stat-label">Tool Calls</div>
                              </div>
                            </div>

                            <div class="agent-meta">
                              <span class="meta-chip">
                                <span class="label">Cost:</span>
                                ${formatCost(agent.usage.estimated_cost_usd)}
                              </span>
                              <span class="meta-chip">
                                <span class="label">Latency:</span>
                                ${formatMs(agent.usage.total_latency_ms)}
                              </span>
                              <span class="meta-chip">
                                <span class="label">Input:</span>
                                ${formatNumber(agent.usage.input_tokens)}
                              </span>
                              <span class="meta-chip">
                                <span class="label">Output:</span>
                                ${formatNumber(agent.usage.output_tokens)}
                              </span>
                            </div>
                          `
                        : html`
                            <div
                              style="color:var(--haira-muted);font-size:0.78rem"
                            >
                              No usage data available.
                            </div>
                          `}
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
    "haira-page-agents": PageAgents;
  }
}
