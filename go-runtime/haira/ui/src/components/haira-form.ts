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
          padding: 2.5rem 1rem 3rem;
        }
        .layout {
          max-width: 960px;
          margin: 0 auto;
          width: 100%;
        }
        @media (min-width: 768px) {
          :host {
            padding: 2.5rem 2rem 3rem;
          }
        }

        /* Header */
        .header {
          margin-bottom: 1.5rem;
        }
        h1 {
          font-size: 1.3rem;
          font-weight: 700;
          color: var(--haira-text);
          display: flex;
          align-items: center;
          gap: 0.6rem;
          margin-bottom: 0.25rem;
        }
        .method-badge {
          font-size: 0.6rem;
          font-weight: 700;
          padding: 0.15rem 0.45rem;
          border-radius: 4px;
          color: #fff;
          letter-spacing: 0.02em;
          flex-shrink: 0;
        }
        .desc {
          font-size: 0.85rem;
          color: var(--haira-muted);
          line-height: 1.45;
        }
        .path {
          font-family: var(--haira-mono);
          font-size: 0.78rem;
          color: var(--haira-muted);
          opacity: 0.7;
          margin-top: 0.15rem;
        }

        /* Form card */
        .card {
          background: var(--haira-bg-card);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          padding: 1.25rem;
          transition: opacity 0.2s;
        }
        .card.disabled {
          opacity: 0.45;
          pointer-events: none;
        }

        /* Submit button */
        .submit-btn {
          width: 100%;
          padding: 0.65rem 1.5rem;
          border: none;
          background: var(--haira-gold);
          color: #0a0a0a;
          border-radius: var(--haira-radius-sm);
          font-size: 0.88rem;
          font-weight: 600;
          cursor: pointer;
          font-family: var(--haira-font);
          transition: all 0.15s;
          margin-top: 0.5rem;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 0.4rem;
        }
        .submit-btn:hover:not(:disabled) {
          background: var(--haira-gold-light);
          box-shadow: 0 2px 16px rgba(232, 163, 23, 0.2);
        }
        .submit-btn:active:not(:disabled) {
          transform: scale(0.99);
        }
        .submit-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }
        .spinner {
          display: inline-block;
          width: 14px;
          height: 14px;
          border: 2px solid #0a0a0a;
          border-top-color: transparent;
          border-radius: 50%;
          animation: spin 0.6s linear infinite;
        }
        @keyframes spin { to { transform: rotate(360deg); } }

        /* Pipeline + result area */
        .output-area {
          margin-top: 0.5rem;
        }
      </style>
      <div class="layout">
        <div class="header">
          <h1>
            ${this.esc(m.title || m.name)}
            <span class="method-badge" style="background:${methodColor(m.method)}">${m.method}</span>
          </h1>
          ${m.description ? `<p class="desc">${this.esc(m.description)}</p>` : ""}
          <p class="path">${this.esc(m.path)}</p>
        </div>
        <div class="card" id="card">
          <form id="wf-form">
            <div id="fields"></div>
            <button type="submit" class="submit-btn" id="submit-btn">Run</button>
          </form>
        </div>
        <div class="output-area" id="output-area"></div>
      </div>
    `;

    // Create fields
    const fieldsContainer = shadow.getElementById("fields")!;
    for (const param of m.params) {
      const field = document.createElement("haira-field") as HairaField;
      field.setAttribute("name", param.Name);
      field.setAttribute("type", param.Type);
      fieldsContainer.appendChild(field);
    }

    // Create pipeline and result
    const outputArea = shadow.getElementById("output-area")!;
    const pipeline = document.createElement("haira-pipeline") as HairaPipeline;
    outputArea.appendChild(pipeline);
    if (m.steps && m.steps.length > 0) {
      pipeline.setSteps(m.steps);
    }

    const result = document.createElement("haira-result") as HairaResult;
    outputArea.appendChild(result);

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
      if (m.steps && m.steps.length > 0) {
        pipeline.reset();
        pipeline.show();

        // For file uploads, ensure all fields are in formData
        let sseFormData: FormData | undefined;
        if (hasFile && formData) {
          for (const [k, v] of Object.entries(body)) {
            if (!formData.has(k)) formData.append(k, String(v));
          }
          sseFormData = formData;
        }

        await streamSSE(
          m.path,
          body,
          {
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
          },
          sseFormData,
        );
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
