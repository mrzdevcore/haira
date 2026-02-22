import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import type { FormViewProps } from "../core/types";

@customElement("haira-ui-form-view")
export class HairaFormView extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      .form-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .form-header {
        padding: 0.55rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .form-title {
        font-size: 0.85rem;
        font-weight: 700;
        color: var(--haira-text);
      }

      .form-body {
        padding: 0.75rem 0.85rem;
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
      }

      .field-group {
        display: flex;
        flex-direction: column;
        gap: 0.3rem;
      }
      .field-label {
        font-size: 0.78rem;
        font-weight: 600;
        color: var(--haira-text-dim);
        display: flex;
        align-items: center;
        gap: 0.3rem;
      }
      .field-required {
        color: var(--haira-error);
        font-size: 0.72rem;
      }

      input[type="text"],
      input[type="number"],
      textarea {
        width: 100%;
        padding: 0.5rem 0.7rem;
        background: var(--haira-bg-input);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        color: var(--haira-text);
        font-family: var(--haira-font);
        font-size: 0.84rem;
        outline: none;
        transition: border-color 0.15s;
      }
      input:focus,
      textarea:focus,
      select:focus {
        border-color: var(--haira-border-focus);
      }
      input::placeholder,
      textarea::placeholder {
        color: var(--haira-muted);
      }
      textarea {
        min-height: 72px;
        resize: vertical;
      }

      select {
        width: 100%;
        padding: 0.5rem 0.7rem;
        background: var(--haira-bg-input);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        color: var(--haira-text);
        font-family: var(--haira-font);
        font-size: 0.84rem;
        outline: none;
        transition: border-color 0.15s;
        cursor: pointer;
        appearance: none;
        -webkit-appearance: none;
        background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1L5 5L9 1' stroke='%2371717a' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
        background-repeat: no-repeat;
        background-position: right 0.65rem center;
        padding-right: 2rem;
      }

      /* Checkbox for bool fields */
      .checkbox-row {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        cursor: pointer;
        user-select: none;
      }
      .checkbox-row input[type="checkbox"] {
        width: 16px;
        height: 16px;
        accent-color: var(--haira-accent);
        cursor: pointer;
      }
      .checkbox-label {
        font-size: 0.84rem;
        color: var(--haira-text);
      }

      .form-actions {
        padding: 0.65rem 0.85rem;
        border-top: 1px solid var(--haira-border);
        display: flex;
        justify-content: flex-end;
      }
      .submit-btn {
        padding: 0.5rem 1.2rem;
        background: var(--haira-accent);
        color: #fff;
        border: none;
        border-radius: var(--haira-radius-sm);
        font-family: var(--haira-font);
        font-size: 0.82rem;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.15s;
      }
      .submit-btn:hover {
        opacity: 0.9;
        transform: translateY(-1px);
      }
      .submit-btn:active {
        transform: translateY(0);
      }
    `,
  ];

  @property({ type: String }) title: string = "";
  @property({ type: Array }) fields: FormViewProps["fields"] = [];
  @property({ type: String }) submit_label: string = "Submit";
  @property({ type: String }) submit_action: string = "";

  /** Set all props at once */
  public setProps(props: FormViewProps): void {
    this.title = props.title || "";
    this.fields = props.fields || [];
    this.submit_label = props.submit_label || "Submit";
    this.submit_action = props.submit_action || "";
  }

  private _collectValues(): Record<string, unknown> {
    const data: Record<string, unknown> = {};
    const inputs = this.renderRoot.querySelectorAll("[data-field]");
    inputs.forEach((el) => {
      const name = el.getAttribute("data-field") || "";
      if (el instanceof HTMLInputElement) {
        if (el.type === "checkbox") {
          data[name] = el.checked;
        } else if (el.type === "number") {
          data[name] = parseFloat(el.value) || 0;
        } else {
          data[name] = el.value;
        }
      } else if (
        el instanceof HTMLTextAreaElement ||
        el instanceof HTMLSelectElement
      ) {
        data[name] = el.value;
      }
    });
    return data;
  }

  private _onSubmit() {
    const values = this._collectValues();
    this.dispatchEvent(
      new CustomEvent("haira-form-submit", {
        bubbles: true,
        composed: true,
        detail: {
          action: this.submit_action,
          values,
        },
      })
    );
  }

  private _renderField(field: FormViewProps["fields"][number]) {
    const label = field.label || field.name;
    const fieldType = field.field_type || "text";

    return html`
      <div class="field-group">
        <label class="field-label">
          ${label}
          ${field.required
            ? html`<span class="field-required">*</span>`
            : nothing}
        </label>
        ${this._renderFieldInput(field, fieldType)}
      </div>
    `;
  }

  private _renderFieldInput(
    field: FormViewProps["fields"][number],
    fieldType: string
  ) {
    // Select / dropdown
    if (field.options && field.options.length > 0) {
      return html`
        <select data-field=${field.name}>
          ${field.options.map(
            (opt) =>
              html`<option
                value=${opt}
                ?selected=${opt === field.value}
              >
                ${opt}
              </option>`
          )}
        </select>
      `;
    }

    switch (fieldType) {
      case "bool":
      case "boolean":
      case "checkbox":
        return html`
          <label class="checkbox-row">
            <input
              type="checkbox"
              data-field=${field.name}
              ?checked=${field.value === "true" || field.value === "1"}
            />
            <span class="checkbox-label">${field.value ? "Yes" : "No"}</span>
          </label>
        `;

      case "number":
      case "int":
      case "float":
        return html`
          <input
            type="number"
            data-field=${field.name}
            .value=${field.value || ""}
            placeholder="Enter a number..."
            step=${fieldType === "int" ? "1" : "any"}
            ?required=${field.required}
          />
        `;

      case "textarea":
      case "text_area":
        return html`
          <textarea
            data-field=${field.name}
            .value=${field.value || ""}
            placeholder="Enter text..."
            ?required=${field.required}
          ></textarea>
        `;

      default:
        return html`
          <input
            type="text"
            data-field=${field.name}
            .value=${field.value || ""}
            placeholder="Enter value..."
            ?required=${field.required}
          />
        `;
    }
  }

  render() {
    return html`
      <div class="form-card">
        ${this.title
          ? html`
              <div class="form-header">
                <span class="form-title">${this.title}</span>
              </div>
            `
          : nothing}
        <div class="form-body">
          ${this.fields.map((f) => this._renderField(f))}
        </div>
        <div class="form-actions">
          <button class="submit-btn" @click=${this._onSubmit}>
            ${this.submit_label}
          </button>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-form-view": HairaFormView;
  }
}
