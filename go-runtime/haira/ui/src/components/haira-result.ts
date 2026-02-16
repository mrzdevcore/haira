import {
  baseStyles,
  sharedKeyframes,
  scrollbarStyles,
  iconCopy,
  iconCopyDone,
} from "../theme";

const analysisStyles = `
  .analysis {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    font-family: var(--haira-font);
    padding: 0.25rem 0;
  }
  .analysis-section {
    border-radius: var(--haira-radius-sm);
    border: 1px solid var(--haira-border);
    overflow: hidden;
  }
  .analysis-section-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    font-weight: 700;
    font-size: 0.7rem;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .analysis-section-header .section-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    border-radius: 4px;
    flex-shrink: 0;
  }
  .analysis-section-body {
    padding: 0.6rem 0.75rem;
    font-size: 0.82rem;
    line-height: 1.65;
    color: var(--haira-text);
    background: var(--haira-bg-card);
  }
  .analysis-section-body p {
    margin: 0;
  }
  .analysis-section-body code {
    background: var(--haira-bg-elevated);
    border: 1px solid var(--haira-border);
    border-radius: 3px;
    padding: 0.1rem 0.35rem;
    font-family: var(--haira-mono);
    font-size: 0.78rem;
    color: var(--haira-gold-light);
  }
  .analysis-section-body ul {
    margin: 0.3rem 0 0 0;
    padding-left: 1.2rem;
  }
  .analysis-section-body li {
    margin-bottom: 0.2rem;
  }

  /* ROOT CAUSE - red tint */
  .section-root-cause .analysis-section-header {
    background: rgba(239, 68, 68, 0.08);
    color: #f87171;
    border-bottom: 1px solid rgba(239, 68, 68, 0.15);
  }
  .section-root-cause .section-icon {
    background: rgba(239, 68, 68, 0.15);
    color: #f87171;
  }

  /* AFFECTED - amber/warn tint */
  .section-affected .analysis-section-header {
    background: rgba(234, 179, 8, 0.08);
    color: #facc15;
    border-bottom: 1px solid rgba(234, 179, 8, 0.15);
  }
  .section-affected .section-icon {
    background: rgba(234, 179, 8, 0.15);
    color: #facc15;
  }

  /* FIX - green tint */
  .section-fix .analysis-section-header {
    background: rgba(34, 197, 94, 0.08);
    color: #4ade80;
    border-bottom: 1px solid rgba(34, 197, 94, 0.15);
  }
  .section-fix .section-icon {
    background: rgba(34, 197, 94, 0.15);
    color: #4ade80;
  }

  /* MISSING VALUES - blue tint */
  .section-missing .analysis-section-header {
    background: rgba(59, 130, 246, 0.08);
    color: #60a5fa;
    border-bottom: 1px solid rgba(59, 130, 246, 0.15);
  }
  .section-missing .section-icon {
    background: rgba(59, 130, 246, 0.15);
    color: #60a5fa;
  }

  /* Generic / unknown section - muted */
  .section-generic .analysis-section-header {
    background: var(--haira-bg-elevated);
    color: var(--haira-text-dim);
    border-bottom: 1px solid var(--haira-border);
  }
  .section-generic .section-icon {
    background: rgba(113, 113, 122, 0.2);
    color: var(--haira-muted);
  }

  .analysis-prefix {
    font-size: 0.8rem;
    color: var(--haira-text-dim);
    padding: 0.25rem 0;
    line-height: 1.5;
    border-bottom: 1px solid var(--haira-border);
    margin-bottom: 0.25rem;
    padding-bottom: 0.6rem;
  }
  .analysis-prefix code {
    background: var(--haira-bg-elevated);
    border: 1px solid var(--haira-border);
    border-radius: 3px;
    padding: 0.1rem 0.35rem;
    font-family: var(--haira-mono);
    font-size: 0.78rem;
    color: var(--haira-error);
  }
`;

const sectionIcons: Record<string, string> = {
  "root-cause": `<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="1.5"/><path d="M8 5v3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><circle cx="8" cy="11" r="0.8" fill="currentColor"/></svg>`,
  affected: `<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M8 2L14 13H2L8 2z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/><path d="M8 7v2.5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/><circle cx="8" cy="11" r="0.6" fill="currentColor"/></svg>`,
  fix: `<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M4.5 8.5L7 11L11.5 5.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
  missing: `<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="3" y="3" width="10" height="10" rx="2" stroke="currentColor" stroke-width="1.5"/><path d="M6 8h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>`,
  generic: `<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="1.5"/><path d="M8 5.5v5M5.5 8h5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg>`,
};

interface AnalysisSection {
  key: string;
  label: string;
  content: string;
}

function classifySection(heading: string): string {
  const h = heading.toUpperCase();
  if (h.includes("ROOT") || h.includes("CAUSE")) return "root-cause";
  if (h.includes("AFFECTED") || h.includes("TABLE") || h.includes("COLUMN"))
    return "affected";
  if (h.includes("FIX") || h.includes("SOLUTION") || h.includes("SUGGEST"))
    return "fix";
  if (h.includes("MISSING") || h.includes("VALUE")) return "missing";
  return "generic";
}

function parseAnalysis(text: string): {
  prefix: string;
  sections: AnalysisSection[];
} | null {
  // Detect structured analysis: look for known section headers
  // Patterns: "ROOT CAUSE:", "**ROOT CAUSE**:", "AFFECTED:", "FIX:", etc.
  const sectionPattern =
    /(?:^|\n)\s*(?:\*{0,2})(ROOT\s*CAUSE|AFFECTED(?:\s+TABLES?(?:\s+AND\s+COLUMNS?)?)?|MISSING\s+(?:FK\s+)?VALUES?|FIX|SUGGESTED?\s+FIX|SOLUTION)(?:\*{0,2})\s*:?\s*/gi;

  const matches = [...text.matchAll(sectionPattern)];
  if (matches.length < 2) return null; // Not structured enough

  const prefix = text.slice(0, matches[0].index!).trim();
  const sections: AnalysisSection[] = [];

  for (let i = 0; i < matches.length; i++) {
    const start = matches[i].index! + matches[i][0].length;
    const end = i + 1 < matches.length ? matches[i + 1].index! : text.length;
    const content = text.slice(start, end).trim();
    const label = matches[i][1].trim().replace(/\*+/g, "");
    const key = classifySection(label);
    sections.push({ key, label, content });
  }

  return { prefix, sections };
}

function formatContent(content: string): string {
  // Convert backtick `code` to <code>
  let html = escapeHtml(content);
  html = html.replace(/`([^`]+)`/g, "<code>$1</code>");

  // Convert lines starting with - or * to list items
  const lines = html.split("\n");
  let result = "";
  let inList = false;

  for (const line of lines) {
    const trimmed = line.trim();
    const isListItem = /^[-*]\s+/.test(trimmed);

    if (isListItem) {
      if (!inList) {
        result += "<ul>";
        inList = true;
      }
      result += `<li>${trimmed.replace(/^[-*]\s+/, "")}</li>`;
    } else {
      if (inList) {
        result += "</ul>";
        inList = false;
      }
      if (trimmed) {
        result += `<p>${trimmed}</p>`;
      }
    }
  }
  if (inList) result += "</ul>";

  return result;
}

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

export class HairaResult extends HTMLElement {
  private rawText = "";

  connectedCallback() {
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host { display: none; margin-top: 0.75rem; }
        :host([visible]) {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 0.6rem 0.85rem;
          border-bottom: 1px solid var(--haira-border);
        }
        .header-left {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          font-weight: 600;
          font-size: 0.78rem;
          color: var(--haira-muted);
        }
        .dot {
          width: 6px;
          height: 6px;
          border-radius: 50%;
          flex-shrink: 0;
        }
        .dot.success { background: var(--haira-success); }
        .dot.error { background: var(--haira-error); }
        .copy-btn {
          background: none;
          border: 1px solid transparent;
          border-radius: 4px;
          padding: 0.25rem;
          cursor: pointer;
          color: var(--haira-muted);
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.15s;
        }
        .copy-btn:hover {
          color: var(--haira-gold);
          border-color: var(--haira-border);
          background: var(--haira-bg-elevated);
        }
        .body {
          padding: 0.85rem;
          font-size: 0.8rem;
          line-height: 1.6;
          word-break: break-word;
          color: var(--haira-text-dim);
          max-height: 500px;
          overflow-y: auto;
          ${scrollbarStyles}
        }
        .body.raw {
          font-family: var(--haira-mono);
          white-space: pre-wrap;
        }
        ${analysisStyles}
      </style>
      <div class="card">
        <div class="header">
          <div class="header-left">
            <span class="dot" id="dot"></span>
            <span id="label">Result</span>
          </div>
          <button class="copy-btn" id="copy-btn" title="Copy to clipboard">${iconCopy}</button>
        </div>
        <div class="body" id="body"></div>
      </div>
    `;

    shadow
      .getElementById("copy-btn")!
      .addEventListener("click", () => this.copyResult());
  }

  show(data: unknown, isError: boolean) {
    this.setAttribute("visible", "");
    const body = this.shadowRoot!.getElementById("body")!;
    const dot = this.shadowRoot!.getElementById("dot")!;
    const label = this.shadowRoot!.getElementById("label")!;
    dot.className = `dot ${isError ? "error" : "success"}`;
    label.textContent = isError ? "Error" : "Result";

    let text: string;
    if (typeof data === "string") {
      text = data;
    } else {
      text = JSON.stringify(data, null, 2);
    }

    this.rawText = text;

    // Try to parse as structured analysis
    const parsed = parseAnalysis(text);
    if (parsed && isError) {
      body.classList.remove("raw");
      body.innerHTML = this.renderAnalysis(parsed);
      label.textContent = "Error Analysis";
    } else {
      body.classList.add("raw");
      body.textContent = text;
    }
  }

  private renderAnalysis(parsed: {
    prefix: string;
    sections: AnalysisSection[];
  }): string {
    let html = '<div class="analysis">';

    if (parsed.prefix) {
      // Format the prefix (usually contains the raw PG error)
      let prefixHtml = escapeHtml(parsed.prefix);
      prefixHtml = prefixHtml.replace(/`([^`]+)`/g, "<code>$1</code>");
      html += `<div class="analysis-prefix">${prefixHtml}</div>`;
    }

    for (const section of parsed.sections) {
      const cssClass = `section-${section.key}`;
      const icon = sectionIcons[section.key] || sectionIcons.generic;
      const displayLabel = section.label
        .split(/\s+/)
        .map((w) => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase())
        .join(" ");

      html += `
        <div class="analysis-section ${cssClass}">
          <div class="analysis-section-header">
            <span class="section-icon">${icon}</span>
            ${escapeHtml(displayLabel)}
          </div>
          <div class="analysis-section-body">
            ${formatContent(section.content)}
          </div>
        </div>`;
    }

    html += "</div>";
    return html;
  }

  hide() {
    this.removeAttribute("visible");
  }

  private async copyResult() {
    const btn = this.shadowRoot?.getElementById("copy-btn");
    if (!btn) return;
    try {
      await navigator.clipboard.writeText(this.rawText);
      btn.innerHTML = iconCopyDone;
      setTimeout(() => {
        btn.innerHTML = iconCopy;
      }, 1500);
    } catch {
      // clipboard API not available
    }
  }
}
