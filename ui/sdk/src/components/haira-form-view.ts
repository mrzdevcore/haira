import { BaseComponent, animateInCSS, cardCSS } from "../core";
import type { FormViewProps } from "../core/types";

export class HairaFormView extends BaseComponent<FormViewProps> {
  protected render() {
    return `
      <div class="card">
        <div class="title-bar" id="title"></div>
        <div class="fields" id="fields"></div>
        <div class="submit-area">
          <button class="submit-btn" id="submit-btn">Submit</button>
        </div>
      </div>`;
  }

  protected styles() {
    return `
      ${animateInCSS}
      .card { ${cardCSS} }
      .title-bar {
        padding: 0.6rem 1rem; font-size: 0.8rem; font-weight: 600;
        color: var(--haira-text); border-bottom: 1px solid var(--haira-border); display: none;
      }
      .fields { padding: 0.75rem 1rem; display: flex; flex-direction: column; gap: 0.6rem; }
      .field-label { font-size: 0.75rem; font-weight: 600; color: var(--haira-text-dim); margin-bottom: 0.2rem; }
      .field-label .required { color: var(--haira-error); margin-left: 0.2rem; }
      input, select, textarea {
        width: 100%; background: var(--haira-bg); border: 1px solid var(--haira-border);
        color: var(--haira-text); padding: 0.45rem 0.65rem;
        border-radius: var(--haira-radius-sm); font-size: 0.8rem;
        font-family: var(--haira-font); outline: none; transition: border-color 0.15s;
      }
      input:focus, select:focus, textarea:focus { border-color: var(--haira-border-focus); }
      textarea { min-height: 60px; resize: vertical; }
      select {
        cursor: pointer; appearance: none;
        background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1l4 4 4-4' stroke='%2371717a' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E");
        background-repeat: no-repeat; background-position: right 0.6rem center; padding-right: 2rem;
      }
      .submit-area { padding: 0.5rem 1rem 0.75rem; border-top: 1px solid var(--haira-border); }
      .submit-btn {
        background: var(--haira-accent); color: #1a0e04; border: none;
        padding: 0.5rem 1.2rem; border-radius: var(--haira-radius-sm);
        font-size: 0.8rem; font-weight: 600; font-family: var(--haira-font);
        cursor: pointer; transition: all 0.15s;
      }
      .submit-btn:hover { background: var(--haira-accent-light); box-shadow: 0 2px 12px rgba(232, 163, 23, 0.25); }`;
  }

  protected onUpdate() {
    const { title, fields = [], submit_label = "Submit", submit_action = "" } = this.props;

    const titleEl = this.$("title");
    if (title) {
      titleEl.textContent = title;
      titleEl.style.display = "";
    }

    this.$("submit-btn").textContent = submit_label;

    const container = this.$("fields");
    container.innerHTML = fields
      .map((field) => {
        const name = field.name || "";
        const label = field.label || name;
        const type = field.field_type || "text";
        const value = field.value || "";
        const required = field.required;
        const options = field.options || [];

        let input: string;
        if (type === "select" && options.length > 0) {
          input = `<select name="${this.escAttr(name)}">
            ${options.map((o) => `<option value="${this.escAttr(o)}" ${o === value ? "selected" : ""}>${this.esc(o)}</option>`).join("")}
          </select>`;
        } else if (type === "textarea") {
          input = `<textarea name="${this.escAttr(name)}">${this.esc(value)}</textarea>`;
        } else {
          input = `<input type="${this.escAttr(type)}" name="${this.escAttr(name)}" value="${this.escAttr(value)}" ${required ? "required" : ""} />`;
        }

        return `<div class="field-group">
          <div class="field-label">${this.esc(label)}${required ? '<span class="required">*</span>' : ""}</div>
          ${input}
        </div>`;
      })
      .join("");

    this.$("submit-btn").onclick = () => {
      const formData: Record<string, string> = {};
      container.querySelectorAll("input, select, textarea").forEach((el) => {
        const input = el as HTMLInputElement;
        formData[input.name] = input.value;
      });
      this.emit("haira-form-submit", { action: submit_action, data: formData });
    };
  }
}
