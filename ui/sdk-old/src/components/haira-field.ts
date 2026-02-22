import { baseCSS, esc } from "../core";

export class HairaField extends HTMLElement {
  private _fileValue: File | null = null;

  connectedCallback() {
    const name = this.getAttribute("name") || "";
    const type = this.getAttribute("type") || "string";

    const shadow = this.attachShadow({ mode: "open" });
    shadow.innerHTML = `
      <style>
        ${baseCSS}
        :host { display: block; margin-bottom: 1rem; }
        .field-label { display: block; font-weight: 500; font-size: 0.82rem; color: var(--haira-text-dim); margin-bottom: 0.4rem; }
        .field-sublabel { font-weight: 400; color: var(--haira-muted); font-size: 0.75rem; margin-left: 0.25rem; }
        input[type="text"], input[type="number"] {
          width: 100%; padding: 0.55rem 0.75rem; background: var(--haira-bg-input);
          border: 1px solid var(--haira-border); border-radius: var(--haira-radius-sm);
          color: var(--haira-text); font-size: 0.88rem; font-family: var(--haira-font);
          outline: none; transition: border-color 0.15s, box-shadow 0.15s;
        }
        input[type="text"]:focus, input[type="number"]:focus {
          border-color: var(--haira-accent); box-shadow: 0 0 0 3px rgba(232, 163, 23, 0.08);
        }
        input[type="text"]::placeholder, input[type="number"]::placeholder { color: var(--haira-muted); opacity: 0.6; }
        .toggle-row { display: flex; align-items: center; justify-content: space-between; padding: 0.5rem 0; }
        .toggle-label { font-size: 0.88rem; font-weight: 500; color: var(--haira-text-dim); }
        .toggle { position: relative; width: 40px; height: 22px; flex-shrink: 0; }
        .toggle input { opacity: 0; width: 0; height: 0; position: absolute; }
        .toggle-track {
          position: absolute; inset: 0; background: var(--haira-bg-elevated);
          border: 1px solid var(--haira-border); border-radius: 11px;
          cursor: pointer; transition: background 0.2s, border-color 0.2s;
        }
        .toggle-track::after {
          content: ""; position: absolute; top: 2px; left: 2px;
          width: 16px; height: 16px; background: var(--haira-muted);
          border-radius: 50%; transition: transform 0.2s ease, background 0.2s;
        }
        .toggle input:checked + .toggle-track { background: rgba(232, 163, 23, 0.15); border-color: var(--haira-accent); }
        .toggle input:checked + .toggle-track::after { transform: translateX(18px); background: var(--haira-accent); }
        .toggle input:focus-visible + .toggle-track { box-shadow: 0 0 0 3px rgba(232, 163, 23, 0.15); }
        .drop-zone {
          position: relative; border: 1.5px dashed var(--haira-border);
          border-radius: var(--haira-radius); padding: 1.25rem 1rem;
          text-align: center; cursor: pointer; transition: all 0.2s; background: var(--haira-bg-input);
        }
        .drop-zone:hover, .drop-zone.dragover { border-color: var(--haira-accent); background: rgba(232, 163, 23, 0.03); }
        .drop-zone.has-file { border-style: solid; border-color: var(--haira-success); background: rgba(34, 197, 94, 0.03); }
        .drop-icon { margin-bottom: 0.35rem; color: var(--haira-muted); }
        .drop-zone.has-file .drop-icon { color: var(--haira-success); }
        .drop-text { font-size: 0.82rem; color: var(--haira-muted); line-height: 1.4; }
        .drop-text strong { color: var(--haira-accent); cursor: pointer; }
        .drop-zone.has-file .drop-text { color: var(--haira-text-dim); }
        .drop-text .filename { display: block; font-family: var(--haira-mono); font-size: 0.8rem; color: var(--haira-text); margin-top: 0.2rem; word-break: break-all; }
        .drop-text .filesize { font-size: 0.72rem; color: var(--haira-muted); }
        .drop-zone input[type="file"] { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }
        .clear-btn {
          display: none; position: absolute; top: 0.5rem; right: 0.5rem;
          background: var(--haira-bg-elevated); border: 1px solid var(--haira-border);
          border-radius: 4px; color: var(--haira-muted); font-size: 0.7rem;
          padding: 0.15rem 0.4rem; cursor: pointer; transition: all 0.15s;
        }
        .clear-btn:hover { color: var(--haira-error); border-color: var(--haira-error); }
        .drop-zone.has-file .clear-btn { display: block; }
      </style>
      ${this.renderInput(name, type)}
    `;

    if (type === "file") {
      this.setupFileDrop(shadow);
    }
  }

  private renderInput(name: string, type: string): string {
    switch (type) {
      case "bool":
        return `
          <div class="toggle-row">
            <span class="toggle-label">${esc(name)}</span>
            <label class="toggle">
              <input type="checkbox" id="f-${name}" name="${name}">
              <span class="toggle-track"></span>
            </label>
          </div>`;
      case "file":
        return `
          <label class="field-label">${esc(name)}</label>
          <div class="drop-zone" id="drop-zone">
            <input type="file" id="f-${name}" name="${name}">
            <div class="drop-icon">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none"><path d="M12 16V4m0 0l-4 4m4-4l4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </div>
            <div class="drop-text" id="drop-text">
              Drop file here or <strong>browse</strong>
            </div>
            <button class="clear-btn" id="clear-btn" type="button">Clear</button>
          </div>`;
      case "int":
        return `
          <label class="field-label" for="f-${name}">${esc(name)} <span class="field-sublabel">integer</span></label>
          <input type="number" id="f-${name}" name="${name}" step="1" placeholder="0">`;
      case "float":
        return `
          <label class="field-label" for="f-${name}">${esc(name)} <span class="field-sublabel">number</span></label>
          <input type="number" id="f-${name}" name="${name}" step="any" placeholder="0.0">`;
      default:
        return `
          <label class="field-label" for="f-${name}">${esc(name)}</label>
          <input type="text" id="f-${name}" name="${name}" placeholder="Enter value...">`;
    }
  }

  private setupFileDrop(shadow: ShadowRoot) {
    const zone = shadow.getElementById("drop-zone");
    const input = shadow.querySelector("input[type=file]") as HTMLInputElement;
    const textEl = shadow.getElementById("drop-text");
    const clearBtn = shadow.getElementById("clear-btn");
    if (!zone || !input || !textEl || !clearBtn) return;

    const setFile = (file: File) => {
      this._fileValue = file;
      zone.classList.add("has-file");
      const size =
        file.size < 1024
          ? `${file.size} B`
          : file.size < 1048576
            ? `${(file.size / 1024).toFixed(1)} KB`
            : `${(file.size / 1048576).toFixed(1)} MB`;
      textEl.innerHTML = `<span class="filename">${esc(file.name)}</span><span class="filesize">${size}</span>`;
    };

    const clearFile = () => {
      this._fileValue = null;
      input.value = "";
      zone.classList.remove("has-file");
      textEl.innerHTML = `Drop file here or <strong>browse</strong>`;
    };

    input.addEventListener("change", () => {
      if (input.files?.[0]) setFile(input.files[0]);
    });

    clearBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      clearFile();
    });

    zone.addEventListener("dragover", (e) => {
      e.preventDefault();
      zone.classList.add("dragover");
    });
    zone.addEventListener("dragleave", () => {
      zone.classList.remove("dragover");
    });
    zone.addEventListener("drop", (e: DragEvent) => {
      e.preventDefault();
      zone.classList.remove("dragover");
      const file = e.dataTransfer?.files?.[0];
      if (file) {
        setFile(file);
        const dt = new DataTransfer();
        dt.items.add(file);
        input.files = dt.files;
      }
    });
  }

  getValue(): { name: string; value: unknown; type: string } {
    const name = this.getAttribute("name") || "";
    const type = this.getAttribute("type") || "string";

    if (type === "bool") {
      const input = this.shadowRoot!.querySelector("input[type=checkbox]") as HTMLInputElement;
      return { name, value: input.checked, type };
    }
    if (type === "file") {
      return {
        name,
        value:
          this._fileValue ||
          this.shadowRoot!.querySelector<HTMLInputElement>("input[type=file]")?.files?.[0] ||
          null,
        type,
      };
    }
    const input = this.shadowRoot!.querySelector("input") as HTMLInputElement;
    if (type === "int" || type === "float") {
      return { name, value: input.value ? Number(input.value) : "", type };
    }
    return { name, value: input.value, type };
  }
}
