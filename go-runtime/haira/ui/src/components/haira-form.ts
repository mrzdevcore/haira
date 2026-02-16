import { baseStyles, methodColor } from "../theme";
import { submitForm, streamSSE } from "../sse";
import type { WorkflowMeta } from "../types";
import type { HairaField } from "./haira-field";
import type { HairaResult } from "./haira-result";
import type { HairaPipeline } from "./haira-pipeline";

export class HairaForm extends HTMLElement {
  private meta!: WorkflowMeta;

  connectedCallback() {
    this.meta = JSON.parse(this.getAttribute("data-meta") || "{}");
    this.render();
  }

  private render() {
    const m = this.meta;
    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        :host {
          display: block;
          padding: 2rem 1rem;
        }
        .header-area {
          max-width: 600px;
          margin: 0 auto;
        }
        .card-area {
          max-width: 600px;
          margin: 0 auto;
        }
        .pipeline-area {
          max-width: 600px;
          margin: 0 auto;
        }
        h1 {
          font-size: 1.35rem;
          font-weight: 600;
          margin-bottom: 0.15rem;
          display: flex;
          align-items: center;
          gap: 0.5rem;
          color: var(--haira-text);
        }
        .badge {
          font-size: 0.65rem;
          font-weight: 700;
          padding: 0.12rem 0.5rem;
          border-radius: 3px;
          color: #fff;
        }
        .desc {
          font-size: 0.85rem;
          color: var(--haira-muted);
          margin-bottom: 0.2rem;
        }
        .path {
          font-family: var(--haira-mono);
          font-size: 0.82rem;
          color: var(--haira-muted);
          margin-bottom: 1.25rem;
        }
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 1.5rem;
          transition: opacity 0.2s;
        }
        .card.disabled {
          opacity: 0.55;
          pointer-events: none;
        }
        button[type="submit"] {
          width: 100%;
          padding: 0.6rem 1.5rem;
          border: none;
          background: linear-gradient(135deg, var(--haira-gold), var(--haira-gold-light));
          color: #1a0e04;
          border-radius: var(--haira-radius);
          font-size: 0.92rem;
          font-weight: 600;
          cursor: pointer;
          font-family: var(--haira-font);
          transition: all 0.2s;
          margin-top: 0.5rem;
        }
        button[type="submit"]:hover {
          box-shadow: 0 0 24px rgba(232, 163, 23, 0.3);
          transform: translateY(-1px);
        }
        button[type="submit"]:disabled {
          opacity: 0.6;
          cursor: not-allowed;
          transform: none;
          box-shadow: none;
        }
        .spinner {
          display: inline-block;
          width: 14px; height: 14px;
          border: 2px solid #1a0e04;
          border-top-color: transparent;
          border-radius: 50%;
          animation: spin 0.6s linear infinite;
          margin-right: 0.4rem;
          vertical-align: middle;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
      </style>
      <div class="header-area">
        <h1>${this.esc(m.title || m.name)}
          <span class="badge" style="background:${methodColor(m.method)}">${m.method}</span>
        </h1>
        ${m.description ? `<p class="desc">${this.esc(m.description)}</p>` : ""}
        <p class="path">${this.esc(m.path)}</p>
      </div>
      <div class="card-area">
        <div class="card" id="card">
          <form id="wf-form">
            <div id="fields"></div>
            <button type="submit" id="submit-btn">Run</button>
          </form>
        </div>
      </div>
      <div class="pipeline-area" id="after-card"></div>
    `;

    // Create fields
    const fieldsContainer = shadow.getElementById("fields")!;
    for (const param of m.params) {
      const field = document.createElement("haira-field") as HairaField;
      field.setAttribute("name", param.Name);
      field.setAttribute("type", param.Type);
      fieldsContainer.appendChild(field);
    }

    // Create pipeline and result programmatically
    const afterCard = shadow.getElementById("after-card")!;
    const pipeline = document.createElement("haira-pipeline") as HairaPipeline;
    afterCard.appendChild(pipeline);
    if (m.steps && m.steps.length > 0) {
      pipeline.setSteps(m.steps);
    }

    const result = document.createElement("haira-result") as HairaResult;
    afterCard.appendChild(result);

    // Submit handler
    const form = shadow.getElementById("wf-form") as HTMLFormElement;
    const btn = shadow.getElementById("submit-btn") as HTMLButtonElement;
    const card = shadow.getElementById("card")!;

    form.addEventListener("submit", async (e) => {
      e.preventDefault();
      btn.disabled = true;
      btn.innerHTML = '<span class="spinner"></span>Running...';
      card.classList.add("disabled");
      result.hide();

      // Collect field values
      const fields = fieldsContainer.querySelectorAll(
        "haira-field",
      ) as NodeListOf<HairaField>;
      const body: Record<string, unknown> = {};
      let formData: FormData | undefined;
      let hasFile = false;

      for (const field of fields) {
        const { name, value, type } = field.getValue();
        if (type === "file" && value) {
          hasFile = true;
          if (!formData) formData = new FormData();
          formData.append(name, value as File);
        } else if (value !== "" && value !== null) {
          body[name] = value;
          if (formData) formData.append(name, String(value));
        }
      }

      const finish = () => {
        btn.disabled = false;
        btn.textContent = "Run";
        card.classList.remove("disabled");
      };

      // If workflow has steps, use SSE for live pipeline
      if (m.steps && m.steps.length > 0 && !hasFile) {
        pipeline.reset();
        pipeline.show();

        await streamSSE(m.path, body, {
          onStep: (event) => {
            pipeline.updateStep(event);
          },
          onResult: (data) => {
            result.show(data, false);
          },
          onError: (error) => {
            result.show({ error }, true);
          },
          onDone: () => {
            pipeline.finalize();
            finish();
          },
        });
      } else {
        // Simple fetch
        try {
          const resp = await submitForm(
            m.path,
            m.method,
            body,
            hasFile,
            formData,
          );
          result.show(resp.data, resp.status >= 400);
        } catch (err) {
          result.show({ error: (err as Error).message }, true);
        }
        finish();
      }
    });
  }

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
