import { baseStyles, sharedKeyframes } from "../theme";

export class HairaFormView extends HTMLElement {
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
        .fields {
          padding: 0.75rem 1rem;
          display: flex;
          flex-direction: column;
          gap: 0.6rem;
        }
        .field-label {
          font-size: 0.75rem;
          font-weight: 600;
          color: var(--haira-text-dim);
          margin-bottom: 0.2rem;
        }
        .field-label .required {
          color: var(--haira-error);
          margin-left: 0.2rem;
        }
        input, select, textarea {
          width: 100%;
          background: var(--haira-bg);
          border: 1px solid var(--haira-border);
          color: var(--haira-text);
          padding: 0.45rem 0.65rem;
          border-radius: var(--haira-radius-sm);
          font-size: 0.8rem;
          font-family: var(--haira-font);
          outline: none;
          transition: border-color 0.15s;
        }
        input:focus, select:focus, textarea:focus {
          border-color: var(--haira-border-focus);
        }
        textarea { min-height: 60px; resize: vertical; }
        select {
          cursor: pointer;
          appearance: none;
          background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1l4 4 4-4' stroke='%2371717a' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E");
          background-repeat: no-repeat;
          background-position: right 0.6rem center;
          padding-right: 2rem;
        }
        .submit-area {
          padding: 0.5rem 1rem 0.75rem;
          border-top: 1px solid var(--haira-border);
        }
        .submit-btn {
          background: var(--haira-gold);
          color: #1a0e04;
          border: none;
          padding: 0.5rem 1.2rem;
          border-radius: var(--haira-radius-sm);
          font-size: 0.8rem;
          font-weight: 600;
          font-family: var(--haira-font);
          cursor: pointer;
          transition: all 0.15s;
        }
        .submit-btn:hover {
          background: var(--haira-gold-light);
          box-shadow: 0 2px 12px rgba(232, 163, 23, 0.25);
        }
      </style>
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="fields" id="fields"></div>
        <div class="submit-area">
          <button class="submit-btn" id="submit-btn">Submit</button>
        </div>
      </div>
    `;
  }

  setProps(props: Record<string, unknown>) {
    try {
      const titleEl = this.shadowRoot!.getElementById("title")!;
      if (props.title) {
        titleEl.textContent = props.title as string;
        titleEl.style.display = "";
      }

      const submitBtn = this.shadowRoot!.getElementById("submit-btn")!;
      submitBtn.textContent = (props.submit_label as string) || "Submit";

      const fields = (props.fields as Array<Record<string, unknown>>) || [];
      const container = this.shadowRoot!.getElementById("fields")!;
      container.innerHTML = fields
        .map((field) => {
          const name = (field.name as string) || "";
          const label = (field.label as string) || name;
          const type = (field.field_type as string) || "text";
          const value = (field.value as string) || "";
          const required = field.required as boolean;
          const options = (field.options as string[]) || [];

          let input: string;
          if (type === "select" && options.length > 0) {
            input = `<select name="${this.esc(name)}">
              ${options.map((o) => `<option value="${this.esc(o)}" ${o === value ? "selected" : ""}>${this.esc(o)}</option>`).join("")}
            </select>`;
          } else if (type === "textarea") {
            input = `<textarea name="${this.esc(name)}">${this.esc(value)}</textarea>`;
          } else {
            input = `<input type="${this.esc(type)}" name="${this.esc(name)}" value="${this.esc(value)}" ${required ? "required" : ""} />`;
          }

          return `<div class="field-group">
            <div class="field-label">${this.esc(label)}${required ? '<span class="required">*</span>' : ""}</div>
            ${input}
          </div>`;
        })
        .join("");

      // Submit dispatches a custom event with form data
      submitBtn.onclick = () => {
        const formData: Record<string, string> = {};
        container.querySelectorAll("input, select, textarea").forEach((el) => {
          const input = el as HTMLInputElement;
          formData[input.name] = input.value;
        });
        this.dispatchEvent(
          new CustomEvent("haira-form-submit", {
            detail: {
              action: (props.submit_action as string) || "",
              data: formData,
            },
            bubbles: true,
            composed: true,
          }),
        );
      };
    } catch {
      // Graceful fallback
    }
  }

  private esc(s: string): string {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
  }
}
