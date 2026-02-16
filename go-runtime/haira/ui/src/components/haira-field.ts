import { baseStyles } from "../theme";

export class HairaField extends HTMLElement {
  connectedCallback() {
    const name = this.getAttribute("name") || "";
    const type = this.getAttribute("type") || "string";

    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseStyles}
        :host { display: block; margin-bottom: 0.85rem; }
        label {
          display: block;
          font-weight: 600;
          font-size: 0.85rem;
          color: var(--haira-text-dim);
          margin-bottom: 0.3rem;
        }
        input[type="text"],
        input[type="number"] {
          width: 100%;
          padding: 0.5rem 0.75rem;
          background: var(--haira-bg-input);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          color: var(--haira-text);
          font-size: 0.88rem;
          font-family: var(--haira-font);
          outline: none;
          transition: border-color 0.15s, box-shadow 0.15s;
        }
        input:focus {
          border-color: var(--haira-gold);
          box-shadow: 0 0 0 2px rgba(232, 163, 23, 0.1);
        }
        input:focus-visible {
          box-shadow: 0 0 0 2px var(--haira-border-focus);
        }
        input[type="file"] {
          width: 100%;
          padding: 0.5rem 0.75rem;
          background: var(--haira-bg-input);
          border: 1px solid var(--haira-border);
          border-radius: var(--haira-radius);
          color: var(--haira-muted);
          font-size: 0.82rem;
          font-family: var(--haira-font);
          transition: border-color 0.15s;
        }
        input[type="file"]::file-selector-button {
          background: var(--haira-bg-card-hover);
          border: 1px solid var(--haira-border);
          border-radius: 4px;
          color: var(--haira-text-dim);
          padding: 0.25rem 0.6rem;
          margin-right: 0.5rem;
          cursor: pointer;
          font-size: 0.8rem;
          transition: all 0.15s;
        }
        input[type="file"]::file-selector-button:hover {
          background: var(--haira-gold-dim);
          border-color: var(--haira-gold);
          color: var(--haira-gold);
        }
        .checkbox-row {
          display: flex;
          align-items: center;
          gap: 0.5rem;
        }
        .checkbox-row input {
          width: 16px;
          height: 16px;
          accent-color: var(--haira-gold);
        }
        .checkbox-row label {
          margin: 0;
          font-weight: 400;
        }
      </style>
      ${this.renderInput(name, type)}
    `;
  }

  private renderInput(name: string, type: string): string {
    switch (type) {
      case "bool":
        return `
          <div class="checkbox-row">
            <input type="checkbox" id="f-${name}" name="${name}">
            <label for="f-${name}">${name}</label>
          </div>`;
      case "file":
        return `
          <label for="f-${name}">${name}</label>
          <input type="file" id="f-${name}" name="${name}">`;
      case "int":
        return `
          <label for="f-${name}">${name}</label>
          <input type="number" id="f-${name}" name="${name}" step="1">`;
      case "float":
        return `
          <label for="f-${name}">${name}</label>
          <input type="number" id="f-${name}" name="${name}" step="any">`;
      default:
        return `
          <label for="f-${name}">${name}</label>
          <input type="text" id="f-${name}" name="${name}">`;
    }
  }

  getValue(): { name: string; value: unknown; type: string } {
    const name = this.getAttribute("name") || "";
    const type = this.getAttribute("type") || "string";
    const input = this.shadowRoot!.querySelector("input") as HTMLInputElement;

    if (type === "bool") {
      return { name, value: input.checked, type };
    }
    if (type === "file") {
      return { name, value: input.files?.[0] || null, type };
    }
    if (type === "int" || type === "float") {
      return { name, value: input.value ? Number(input.value) : "", type };
    }
    return { name, value: input.value, type };
  }
}
