import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import { iconStrings } from "../core/icons";
import { formatBytes } from "../core/utils";

export interface FieldValue {
  name: string;
  value: string | number | boolean | File | null;
  type: string;
}

@customElement("haira-field")
export class HairaField extends LitElement {
  static styles = [
    baseStyles,
    css`
      :host {
        display: block;
      }

      label {
        display: block;
        font-size: 0.78rem;
        font-weight: 600;
        color: var(--haira-text-dim);
        margin-bottom: 0.35rem;
        text-transform: capitalize;
      }

      input[type="text"],
      input[type="number"] {
        width: 100%;
        padding: 0.55rem 0.75rem;
        background: var(--haira-bg-input);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        color: var(--haira-text);
        font-family: var(--haira-font);
        font-size: 0.85rem;
        outline: none;
        transition: border-color 0.15s;
      }
      input[type="text"]:focus,
      input[type="number"]:focus {
        border-color: var(--haira-border-focus);
      }
      input[type="text"]::placeholder,
      input[type="number"]::placeholder {
        color: var(--haira-muted);
      }

      /* Toggle switch for bool */
      .toggle-row {
        display: flex;
        align-items: center;
        gap: 0.6rem;
        cursor: pointer;
        user-select: none;
      }
      .toggle-track {
        width: 36px;
        height: 20px;
        border-radius: 10px;
        background: var(--haira-bg-elevated);
        border: 1px solid var(--haira-border);
        position: relative;
        transition: background 0.15s, border-color 0.15s;
      }
      .toggle-track.on {
        background: var(--haira-accent);
        border-color: var(--haira-accent);
      }
      .toggle-thumb {
        position: absolute;
        top: 2px;
        left: 2px;
        width: 14px;
        height: 14px;
        border-radius: 50%;
        background: var(--haira-text);
        transition: transform 0.15s;
      }
      .toggle-track.on .toggle-thumb {
        transform: translateX(16px);
        background: #fff;
      }
      .toggle-label {
        font-size: 0.82rem;
        color: var(--haira-text-dim);
      }

      /* File drop zone */
      .drop-zone {
        border: 2px dashed var(--haira-border);
        border-radius: var(--haira-radius);
        padding: 1.5rem 1rem;
        text-align: center;
        cursor: pointer;
        transition: all 0.15s;
        background: var(--haira-bg-input);
      }
      .drop-zone.dragover {
        border-color: var(--haira-accent);
        background: var(--haira-accent-dim);
      }
      .drop-zone-icon {
        display: flex;
        justify-content: center;
        margin-bottom: 0.5rem;
        color: var(--haira-muted);
      }
      .drop-zone-text {
        font-size: 0.82rem;
        color: var(--haira-muted);
      }
      .drop-zone-text strong {
        color: var(--haira-accent);
      }

      /* Selected file display */
      .file-selected {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 0.75rem;
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius-sm);
        margin-top: 0.4rem;
      }
      .file-selected .file-icon {
        display: flex;
        align-items: center;
        color: var(--haira-accent);
        flex-shrink: 0;
      }
      .file-info {
        flex: 1;
        min-width: 0;
      }
      .file-name {
        font-size: 0.82rem;
        font-weight: 600;
        color: var(--haira-text);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .file-size {
        font-size: 0.7rem;
        color: var(--haira-muted);
      }
      .file-remove {
        display: flex;
        align-items: center;
        background: none;
        border: none;
        color: var(--haira-muted);
        cursor: pointer;
        padding: 0.15rem;
        border-radius: 4px;
        transition: all 0.12s;
      }
      .file-remove:hover {
        background: var(--haira-bg-elevated);
        color: var(--haira-error);
      }

      /* Hidden file input */
      input[type="file"] {
        display: none;
      }
    `,
  ];

  @property({ type: String }) name: string = "";
  @property({ type: String, reflect: true }) type: string = "string";

  @state() private _textValue: string = "";
  @state() private _numValue: number = 0;
  @state() private _boolValue: boolean = false;
  @state() private _fileValue: File | null = null;
  @state() private _dragover: boolean = false;

  /** Return the raw field value (string | number | boolean | File | null). */
  public getValue(): string | number | boolean | File | null {
    switch (this.type) {
      case "bool":
        return this._boolValue;
      case "int":
      case "float":
        return this._numValue;
      case "file":
        return this._fileValue;
      default:
        return this._textValue;
    }
  }

  private _onTextInput(e: Event) {
    this._textValue = (e.target as HTMLInputElement).value;
  }

  private _onNumberInput(e: Event) {
    this._numValue = parseFloat((e.target as HTMLInputElement).value) || 0;
  }

  private _onToggle() {
    this._boolValue = !this._boolValue;
  }

  // File handling
  private _onDragOver(e: DragEvent) {
    e.preventDefault();
    this._dragover = true;
  }

  private _onDragLeave() {
    this._dragover = false;
  }

  private _onDrop(e: DragEvent) {
    e.preventDefault();
    this._dragover = false;
    const files = e.dataTransfer?.files;
    if (files && files.length > 0) {
      this._fileValue = files[0];
    }
  }

  private _onFileClick() {
    const input = this.renderRoot.querySelector(
      'input[type="file"]'
    ) as HTMLInputElement | null;
    input?.click();
  }

  private _onFileChange(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this._fileValue = input.files[0];
    }
  }

  private _removeFile() {
    this._fileValue = null;
  }

  render() {
    return html`
      <label>${this.name}</label>
      ${this._renderInput()}
    `;
  }

  private _renderInput() {
    switch (this.type) {
      case "bool":
        return html`
          <div class="toggle-row" @click=${this._onToggle}>
            <div class="toggle-track ${this._boolValue ? "on" : ""}">
              <div class="toggle-thumb"></div>
            </div>
            <span class="toggle-label"
              >${this._boolValue ? "True" : "False"}</span
            >
          </div>
        `;

      case "int":
        return html`
          <input
            type="number"
            step="1"
            .value=${String(this._numValue)}
            placeholder="Enter integer..."
            @input=${this._onNumberInput}
          />
        `;

      case "float":
        return html`
          <input
            type="number"
            step="any"
            .value=${String(this._numValue)}
            placeholder="Enter number..."
            @input=${this._onNumberInput}
          />
        `;

      case "file":
        return html`
          <div
            class="drop-zone ${this._dragover ? "dragover" : ""}"
            @dragover=${this._onDragOver}
            @dragleave=${this._onDragLeave}
            @drop=${this._onDrop}
            @click=${this._onFileClick}
          >
            <div class="drop-zone-icon">${unsafeHTML(iconStrings.attach)}</div>
            <div class="drop-zone-text">
              Drop a file here or <strong>click to browse</strong>
            </div>
          </div>
          <input type="file" @change=${this._onFileChange} />
          ${this._fileValue
            ? html`
                <div class="file-selected">
                  <span class="file-icon"
                    >${unsafeHTML(iconStrings.file)}</span
                  >
                  <div class="file-info">
                    <div class="file-name">${this._fileValue.name}</div>
                    <div class="file-size">
                      ${formatBytes(this._fileValue.size)}
                    </div>
                  </div>
                  <button class="file-remove" @click=${this._removeFile}>
                    ${unsafeHTML(iconStrings.xSmall)}
                  </button>
                </div>
              `
            : nothing}
        `;

      default:
        // string
        return html`
          <input
            type="text"
            .value=${this._textValue}
            placeholder="Enter value..."
            @input=${this._onTextInput}
          />
        `;
    }
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-field": HairaField;
  }
}
