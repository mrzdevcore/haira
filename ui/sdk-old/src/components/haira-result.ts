import { baseCSS, scrollbarCSS, icons, esc } from "../core";

export class HairaResult extends HTMLElement {
  private rawText = "";

  connectedCallback() {
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseCSS}
        :host { display: none; margin-top: 0.75rem; }
        :host([visible]) { display: block; animation: fadeSlideUp 0.25s ease-out; }
        .card {
          background: var(--haira-bg-card); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius); overflow: hidden;
        }
        .header {
          display: flex; align-items: center; justify-content: space-between;
          padding: 0.6rem 0.85rem; border-bottom: 1px solid var(--haira-border);
        }
        .header-left { display: flex; align-items: center; gap: 0.4rem; font-weight: 600; font-size: 0.78rem; color: var(--haira-muted); }
        .dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
        .dot.success { background: var(--haira-success); }
        .dot.error { background: var(--haira-error); }
        .copy-btn {
          background: none; border: 1px solid transparent; border-radius: 4px;
          padding: 0.25rem; cursor: pointer; color: var(--haira-muted);
          display: flex; align-items: center; justify-content: center; transition: all 0.15s;
        }
        .copy-btn:hover { color: var(--haira-accent); border-color: var(--haira-border); background: var(--haira-bg-elevated); }
        .body {
          padding: 0.85rem; font-size: 0.82rem; line-height: 1.6;
          color: var(--haira-text-dim); max-height: 600px; overflow-y: auto; ${scrollbarCSS}
        }
        .body.rich { white-space: normal; word-break: break-word; font-family: var(--haira-font); }
        .body.raw { white-space: pre-wrap; word-break: break-word; font-family: var(--haira-mono); font-size: 0.8rem; }
        .result-section { margin-bottom: 0.75rem; }
        .result-section:last-child { margin-bottom: 0; }
        .section-label {
          font-size: 0.68rem; font-weight: 700; text-transform: uppercase;
          letter-spacing: 0.05em; color: var(--haira-muted); margin-bottom: 0.3rem;
        }
        .section-label.error { color: var(--haira-error); }
        .section-value { color: var(--haira-text); line-height: 1.55; }
        .section-value ul { margin: 0.25rem 0 0 0; padding-left: 1.25rem; }
        .section-value li { margin-bottom: 0.15rem; font-family: var(--haira-mono); font-size: 0.78rem; color: var(--haira-text-dim); }
        .code-block {
          background: var(--haira-bg); border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius-sm); padding: 0.65rem 0.85rem; margin-top: 0.35rem;
          font-family: var(--haira-mono); font-size: 0.75rem; line-height: 1.5;
          white-space: pre-wrap; word-break: break-all; overflow-x: auto; color: var(--haira-text);
        }
        .code-lang { font-size: 0.62rem; text-transform: uppercase; color: var(--haira-muted); letter-spacing: 0.04em; margin-bottom: 0.2rem; font-weight: 600; }
        .result-kv { display: flex; gap: 0.5rem; padding: 0.2rem 0; }
        .result-kv .kv-key { font-size: 0.72rem; font-weight: 600; color: var(--haira-muted); min-width: 60px; flex-shrink: 0; }
        .result-kv .kv-val { color: var(--haira-text); font-size: 0.82rem; }
      </style>
      <div class="card">
        <div class="header">
          <div class="header-left">
            <span class="dot" id="dot"></span>
            <span id="label">Result</span>
          </div>
          <button class="copy-btn" id="copy-btn" title="Copy to clipboard">${icons.copy}</button>
        </div>
        <div class="body raw" id="body"></div>
      </div>
    `;

    shadow.getElementById("copy-btn")!.addEventListener("click", () => this.copyResult());
  }

  show(data: unknown, isError: boolean) {
    this.setAttribute("visible", "");
    const body = this.shadowRoot!.getElementById("body")!;
    const dot = this.shadowRoot!.getElementById("dot")!;
    const label = this.shadowRoot!.getElementById("label")!;
    dot.className = `dot ${isError ? "error" : "success"}`;
    label.textContent = isError ? "Error" : "Result";

    const obj = data as Record<string, unknown>;
    if (
      typeof data === "object" &&
      data !== null &&
      typeof obj.message === "string" &&
      obj.message.length > 0
    ) {
      this.rawText = obj.message as string;
      body.className = "body rich";
      body.innerHTML = this.renderMessage(obj.message as string, obj.status as string | undefined);
      return;
    }

    if (typeof data === "object" && data !== null && !Array.isArray(data)) {
      const keys = Object.keys(obj);
      if (keys.length > 0 && keys.length <= 10 && keys.every((k) => typeof obj[k] !== "object" || obj[k] === null)) {
        this.rawText = JSON.stringify(data, null, 2);
        body.className = "body rich";
        body.innerHTML = keys
          .map(
            (k) =>
              `<div class="result-kv"><span class="kv-key">${esc(k)}</span><span class="kv-val">${esc(String(obj[k] ?? ""))}</span></div>`,
          )
          .join("");
        return;
      }
    }

    let text: string;
    if (typeof data === "string") {
      text = data;
    } else {
      text = JSON.stringify(data, null, 2);
    }
    this.rawText = text;
    body.className = "body raw";
    body.textContent = text;
  }

  hide() {
    this.removeAttribute("visible");
  }

  private renderMessage(message: string, status?: string): string {
    const lines = message.split("\n");
    const sections: Array<{
      type: "heading" | "text" | "list" | "code";
      label?: string;
      lang?: string;
      content: string;
    }> = [];

    let i = 0;
    while (i < lines.length) {
      const line = lines[i];

      const codeMatch = line.match(/^```(\w*)$/);
      if (codeMatch) {
        const lang = codeMatch[1] || "";
        const codeLines: string[] = [];
        i++;
        while (i < lines.length && !lines[i].startsWith("```")) {
          codeLines.push(lines[i]);
          i++;
        }
        i++;
        sections.push({ type: "code", lang, content: codeLines.join("\n") });
        continue;
      }

      const headingMatch = line.match(/^([A-Z][A-Z _]{2,}):(.*)$/);
      if (headingMatch) {
        const label = headingMatch[1].trim();
        const rest = headingMatch[2].trim();
        const contentLines: string[] = rest ? [rest] : [];
        i++;
        while (i < lines.length) {
          const nextLine = lines[i];
          if (nextLine.match(/^[A-Z][A-Z _]{2,}:/) || nextLine.startsWith("```")) {
            break;
          }
          contentLines.push(nextLine);
          i++;
        }
        const content = contentLines.join("\n").trim();
        const contentLinesArr = content.split("\n");
        const isList = contentLinesArr.length > 1 && contentLinesArr.every((l) => l.startsWith("- ") || l.trim() === "");
        if (isList) {
          sections.push({ type: "list", label, content });
        } else {
          sections.push({ type: "heading", label, content });
        }
        continue;
      }

      if (line.trim()) {
        sections.push({ type: "text", content: line });
      }
      i++;
    }

    if (sections.length === 0) {
      return `<div class="section-value">${esc(message)}</div>`;
    }

    return sections
      .map((s) => {
        switch (s.type) {
          case "heading":
            return `<div class="result-section">
              <div class="section-label${status === "error" && s.label?.includes("CAUSE") ? " error" : ""}">${esc(s.label || "")}</div>
              <div class="section-value">${esc(s.content)}</div>
            </div>`;
          case "list":
            return `<div class="result-section">
              <div class="section-label">${esc(s.label || "")}</div>
              <div class="section-value"><ul>${s.content
                .split("\n")
                .filter((l) => l.startsWith("- "))
                .map((l) => `<li>${esc(l.slice(2))}</li>`)
                .join("")}</ul></div>
            </div>`;
          case "code":
            return `<div class="result-section">
              ${s.lang ? `<div class="code-lang">${esc(s.lang)}</div>` : ""}
              <div class="code-block">${esc(s.content)}</div>
            </div>`;
          case "text":
            return `<div class="section-value">${esc(s.content)}</div>`;
          default:
            return "";
        }
      })
      .join("");
  }

  private async copyResult() {
    const btn = this.shadowRoot?.getElementById("copy-btn");
    if (!btn) return;
    try {
      await navigator.clipboard.writeText(this.rawText);
      btn.innerHTML = icons.copyDone;
      setTimeout(() => {
        btn.innerHTML = icons.copy;
      }, 1500);
    } catch {
      /* clipboard API not available */
    }
  }
}
