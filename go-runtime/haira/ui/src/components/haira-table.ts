import { baseStyles, sharedKeyframes, scrollbarStyles } from "../theme";

const ROW_LIMIT = 50;

export class HairaTable extends HTMLElement {
  private allRows: string[][] = [];
  private expanded = false;

  connectedCallback() {
    this.attachShadow({ mode: "open" });
    this.shadowRoot!.innerHTML = `
      <style>
        ${baseStyles}
        ${sharedKeyframes}
        :host {
          display: block;
          animation: fadeSlideUp 0.25s ease-out;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          overflow: hidden;
        }
        .title-bar {
          padding: 0.6rem 1rem;
          font-size: 0.8rem;
          font-weight: 600;
          color: var(--haira-text);
          border-bottom: 1px solid var(--haira-border);
          display: none;
        }
        .table-wrap {
          overflow-x: auto;
          ${scrollbarStyles}
        }
        table {
          width: 100%;
          border-collapse: collapse;
          font-size: 0.78rem;
        }
        th {
          text-align: left;
          padding: 0.5rem 0.75rem;
          font-weight: 600;
          font-size: 0.72rem;
          color: var(--haira-muted);
          text-transform: uppercase;
          letter-spacing: 0.04em;
          background: var(--haira-bg);
          border-bottom: 1px solid var(--haira-border);
          white-space: nowrap;
        }
        td {
          padding: 0.45rem 0.75rem;
          color: var(--haira-text-dim);
          border-bottom: 1px solid var(--haira-border);
          line-height: 1.4;
        }
        tr:last-child td { border-bottom: none; }
        tr:hover td { background: var(--haira-bg-card-hover); }
        tr.highlight td { background: rgba(232, 163, 23, 0.06); }
        .show-more {
          display: none;
          text-align: center;
          padding: 0.5rem;
          border-top: 1px solid var(--haira-border);
        }
        .show-more button {
          background: none;
          border: 1px solid var(--haira-border);
          color: var(--haira-text-dim);
          font-size: 0.75rem;
          font-family: var(--haira-font);
          padding: 0.3rem 0.8rem;
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.15s;
        }
        .show-more button:hover {
          border-color: var(--haira-gold);
          color: var(--haira-gold);
        }
      </style>
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="table-wrap">
          <table>
            <thead id="thead"></thead>
            <tbody id="tbody"></tbody>
          </table>
        </div>
        <div class="show-more" id="show-more">
          <button id="show-more-btn"></button>
        </div>
      </div>
    `;

    this.shadowRoot!.getElementById("show-more-btn")!.addEventListener("click", () => {
      this.expanded = !this.expanded;
      this.renderRows();
    });
  }

  setProps(props: Record<string, unknown>) {
    try {
      const title = props.title as string | undefined;
      const titleEl = this.shadowRoot!.getElementById("title")!;
      if (title) {
        titleEl.textContent = title;
        titleEl.style.display = "";
      }

      const headers = (props.headers as string[]) || [];
      const thead = this.shadowRoot!.getElementById("thead")!;
      thead.innerHTML = `<tr>${headers.map((h) => `<th>${this.esc(h)}</th>`).join("")}</tr>`;

      this.allRows = (props.rows as string[][]) || [];
      const highlight = new Set((props.highlight as number[]) || []);
      this.renderRows(highlight);
    } catch {
      // Graceful fallback
    }
  }

  private renderRows(highlight?: Set<number>) {
    const tbody = this.shadowRoot!.getElementById("tbody")!;
    const showMore = this.shadowRoot!.getElementById("show-more")!;
    const showMoreBtn = this.shadowRoot!.getElementById("show-more-btn")!;

    const rows = this.expanded ? this.allRows : this.allRows.slice(0, ROW_LIMIT);
    tbody.innerHTML = rows
      .map(
        (row, i) =>
          `<tr class="${highlight?.has(i) ? "highlight" : ""}">${row.map((c) => `<td>${this.esc(String(c))}</td>`).join("")}</tr>`,
      )
      .join("");

    if (this.allRows.length > ROW_LIMIT) {
      showMore.style.display = "";
      showMoreBtn.textContent = this.expanded
        ? "Show less"
        : `Show all ${this.allRows.length} rows`;
    } else {
      showMore.style.display = "none";
    }
  }

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
