import { LitElement, html, css, nothing } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import { baseStyles, keyframes, animateInStyles } from "../core/styles";
import type { ChartProps, ChartDataset } from "../core/types";

/** Default color palette starting with Haira Gold */
const DEFAULT_COLORS = [
  "#e8a317",
  "#3b82f6",
  "#22c55e",
  "#ef4444",
  "#a855f7",
  "#06b6d4",
  "#f97316",
  "#ec4899",
  "#14b8a6",
  "#8b5cf6",
];

@customElement("haira-ui-chart")
export class HairaChart extends LitElement {
  static styles = [
    baseStyles,
    animateInStyles,
    css`
      .chart-card {
        background: var(--haira-bg-card);
        border: 1px solid var(--haira-border);
        border-radius: var(--haira-radius);
        overflow: hidden;
      }

      .chart-header {
        padding: 0.55rem 0.85rem;
        border-bottom: 1px solid var(--haira-border);
        background: var(--haira-bg);
      }
      .chart-title {
        font-size: 0.85rem;
        font-weight: 700;
        color: var(--haira-text);
      }

      .chart-body {
        padding: 0.75rem;
        display: flex;
        justify-content: center;
      }

      canvas {
        display: block;
        max-width: 100%;
      }

      .legend {
        display: flex;
        flex-wrap: wrap;
        gap: 0.55rem;
        padding: 0.5rem 0.85rem;
        border-top: 1px solid var(--haira-border);
        justify-content: center;
      }
      .legend-item {
        display: flex;
        align-items: center;
        gap: 0.3rem;
        font-size: 0.74rem;
        color: var(--haira-text-dim);
      }
      .legend-dot {
        width: 10px;
        height: 10px;
        border-radius: 2px;
        flex-shrink: 0;
      }
    `,
  ];

  @property({ type: String }) type: ChartProps["type"] = "line";
  @property({ type: String }) title: string = "";
  @property({ type: Array }) labels: string[] = [];
  @property({ type: Array }) datasets: ChartDataset[] = [];
  @property({ type: Number }) height: number = 300;

  private _canvas: HTMLCanvasElement | null = null;

  /** Set all props at once */
  public setProps(props: ChartProps): void {
    this.type = props.type || "line";
    this.title = props.title || "";
    this.labels = props.labels || [];
    this.datasets = props.datasets || [];
    this.height = props.height || 300;
  }

  protected updated(): void {
    this._canvas = this.renderRoot.querySelector("canvas");
    if (this._canvas) {
      this._drawChart();
    }
  }

  private _getColor(index: number): string {
    return DEFAULT_COLORS[index % DEFAULT_COLORS.length];
  }

  private _datasetColor(ds: ChartDataset, index: number): string {
    return ds.color || this._getColor(index);
  }

  private _drawChart(): void {
    const canvas = this._canvas!;
    const dpr = window.devicePixelRatio || 1;
    const width = canvas.parentElement?.clientWidth || 600;
    const h = this.height;

    canvas.width = width * dpr;
    canvas.height = h * dpr;
    canvas.style.width = `${width}px`;
    canvas.style.height = `${h}px`;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    ctx.scale(dpr, dpr);
    ctx.clearRect(0, 0, width, h);

    switch (this.type) {
      case "pie":
        this._drawPie(ctx, width, h);
        break;
      case "bar":
        this._drawBar(ctx, width, h);
        break;
      case "scatter":
        this._drawScatter(ctx, width, h);
        break;
      case "area":
        this._drawArea(ctx, width, h);
        break;
      case "line":
      default:
        this._drawLine(ctx, width, h);
        break;
    }
  }

  // Shared helper: compute chart area paddings
  private _chartArea(w: number, h: number) {
    return {
      left: 55,
      right: 20,
      top: 15,
      bottom: 35,
      plotW: w - 55 - 20,
      plotH: h - 15 - 35,
    };
  }

  // Get text color from CSS variable
  private _textColor(): string {
    const style = getComputedStyle(this);
    return style.getPropertyValue("--haira-text-dim").trim() || "#a1a1aa";
  }

  private _gridColor(): string {
    return "rgba(63,63,70,0.25)";
  }

  // Compute Y min/max from all datasets
  private _yBounds(): { min: number; max: number } {
    let min = Infinity;
    let max = -Infinity;
    for (const ds of this.datasets) {
      for (const v of ds.data) {
        if (v < min) min = v;
        if (v > max) max = v;
      }
    }
    if (min === Infinity) {
      min = 0;
      max = 100;
    }
    // Add padding
    const range = max - min || 1;
    min = min - range * 0.05;
    max = max + range * 0.05;
    if (min < 0 && this.datasets.every((d) => d.data.every((v) => v >= 0))) {
      min = 0;
    }
    return { min, max };
  }

  // Draw grid lines and Y-axis labels
  private _drawGrid(
    ctx: CanvasRenderingContext2D,
    area: ReturnType<typeof this._chartArea>,
    yBounds: { min: number; max: number },
    steps: number = 5
  ): void {
    const textCol = this._textColor();
    const gridCol = this._gridColor();

    ctx.font = "11px -apple-system, BlinkMacSystemFont, sans-serif";
    ctx.textAlign = "right";
    ctx.textBaseline = "middle";

    for (let i = 0; i <= steps; i++) {
      const frac = i / steps;
      const y = area.top + area.plotH - frac * area.plotH;
      const val = yBounds.min + frac * (yBounds.max - yBounds.min);

      // Grid line
      ctx.strokeStyle = gridCol;
      ctx.lineWidth = 0.5;
      ctx.beginPath();
      ctx.moveTo(area.left, y);
      ctx.lineTo(area.left + area.plotW, y);
      ctx.stroke();

      // Y label
      ctx.fillStyle = textCol;
      const label =
        Math.abs(val) >= 1000
          ? `${(val / 1000).toFixed(1)}k`
          : val % 1 === 0
            ? String(Math.round(val))
            : val.toFixed(1);
      ctx.fillText(label, area.left - 8, y);
    }
  }

  // Draw X-axis labels
  private _drawXLabels(
    ctx: CanvasRenderingContext2D,
    area: ReturnType<typeof this._chartArea>
  ): void {
    const textCol = this._textColor();
    const n = this.labels.length;
    if (n === 0) return;

    ctx.font = "11px -apple-system, BlinkMacSystemFont, sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "top";
    ctx.fillStyle = textCol;

    // Show at most ~10 labels to avoid overlap
    const step = Math.max(1, Math.ceil(n / 10));
    for (let i = 0; i < n; i += step) {
      const x = area.left + (i / Math.max(n - 1, 1)) * area.plotW;
      ctx.fillText(this.labels[i], x, area.top + area.plotH + 10);
    }
  }

  // --- Chart renderers ---

  private _drawLine(
    ctx: CanvasRenderingContext2D,
    w: number,
    h: number
  ): void {
    const area = this._chartArea(w, h);
    const yB = this._yBounds();
    this._drawGrid(ctx, area, yB);
    this._drawXLabels(ctx, area);

    const n = this.labels.length;
    if (n === 0) return;

    this.datasets.forEach((ds, di) => {
      const color = this._datasetColor(ds, di);
      ctx.strokeStyle = color;
      ctx.lineWidth = 2;
      ctx.lineJoin = "round";
      ctx.lineCap = "round";
      ctx.beginPath();

      for (let i = 0; i < ds.data.length && i < n; i++) {
        const x = area.left + (i / Math.max(n - 1, 1)) * area.plotW;
        const yFrac = (ds.data[i] - yB.min) / (yB.max - yB.min);
        const y = area.top + area.plotH - yFrac * area.plotH;
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      }
      ctx.stroke();

      // Data points
      for (let i = 0; i < ds.data.length && i < n; i++) {
        const x = area.left + (i / Math.max(n - 1, 1)) * area.plotW;
        const yFrac = (ds.data[i] - yB.min) / (yB.max - yB.min);
        const y = area.top + area.plotH - yFrac * area.plotH;
        ctx.fillStyle = color;
        ctx.beginPath();
        ctx.arc(x, y, 3, 0, Math.PI * 2);
        ctx.fill();
      }
    });
  }

  private _drawArea(
    ctx: CanvasRenderingContext2D,
    w: number,
    h: number
  ): void {
    const area = this._chartArea(w, h);
    const yB = this._yBounds();
    this._drawGrid(ctx, area, yB);
    this._drawXLabels(ctx, area);

    const n = this.labels.length;
    if (n === 0) return;
    const baseline = area.top + area.plotH;

    this.datasets.forEach((ds, di) => {
      const color = this._datasetColor(ds, di);

      // Fill area
      ctx.beginPath();
      for (let i = 0; i < ds.data.length && i < n; i++) {
        const x = area.left + (i / Math.max(n - 1, 1)) * area.plotW;
        const yFrac = (ds.data[i] - yB.min) / (yB.max - yB.min);
        const y = area.top + area.plotH - yFrac * area.plotH;
        if (i === 0) {
          ctx.moveTo(x, baseline);
          ctx.lineTo(x, y);
        } else {
          ctx.lineTo(x, y);
        }
      }
      const lastX =
        area.left +
        ((Math.min(ds.data.length, n) - 1) / Math.max(n - 1, 1)) *
          area.plotW;
      ctx.lineTo(lastX, baseline);
      ctx.closePath();
      ctx.fillStyle = color + "22"; // alpha
      ctx.fill();

      // Stroke line
      ctx.strokeStyle = color;
      ctx.lineWidth = 2;
      ctx.lineJoin = "round";
      ctx.beginPath();
      for (let i = 0; i < ds.data.length && i < n; i++) {
        const x = area.left + (i / Math.max(n - 1, 1)) * area.plotW;
        const yFrac = (ds.data[i] - yB.min) / (yB.max - yB.min);
        const y = area.top + area.plotH - yFrac * area.plotH;
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      }
      ctx.stroke();
    });
  }

  private _drawBar(
    ctx: CanvasRenderingContext2D,
    w: number,
    h: number
  ): void {
    const area = this._chartArea(w, h);
    const yB = this._yBounds();
    // Ensure 0 is always in view for bars
    if (yB.min > 0) yB.min = 0;
    this._drawGrid(ctx, area, yB);
    this._drawXLabels(ctx, area);

    const n = this.labels.length;
    if (n === 0) return;
    const dsCount = this.datasets.length;
    const groupWidth = area.plotW / n;
    const barWidth = (groupWidth * 0.7) / dsCount;
    const gap = groupWidth * 0.15;

    this.datasets.forEach((ds, di) => {
      const color = this._datasetColor(ds, di);
      ctx.fillStyle = color;

      for (let i = 0; i < ds.data.length && i < n; i++) {
        const groupX = area.left + i * groupWidth;
        const x = groupX + gap + di * barWidth;
        const yFrac = (ds.data[i] - yB.min) / (yB.max - yB.min);
        const barH = yFrac * area.plotH;
        const y = area.top + area.plotH - barH;

        // Rounded top corners
        const r = Math.min(3, barWidth / 2);
        ctx.beginPath();
        ctx.moveTo(x, area.top + area.plotH);
        ctx.lineTo(x, y + r);
        ctx.arcTo(x, y, x + r, y, r);
        ctx.arcTo(x + barWidth, y, x + barWidth, y + r, r);
        ctx.lineTo(x + barWidth, area.top + area.plotH);
        ctx.closePath();
        ctx.fill();
      }
    });
  }

  private _drawScatter(
    ctx: CanvasRenderingContext2D,
    w: number,
    h: number
  ): void {
    const area = this._chartArea(w, h);
    const yB = this._yBounds();
    this._drawGrid(ctx, area, yB);
    this._drawXLabels(ctx, area);

    const n = this.labels.length;
    if (n === 0) return;

    this.datasets.forEach((ds, di) => {
      const color = this._datasetColor(ds, di);
      ctx.fillStyle = color;

      for (let i = 0; i < ds.data.length && i < n; i++) {
        const x = area.left + (i / Math.max(n - 1, 1)) * area.plotW;
        const yFrac = (ds.data[i] - yB.min) / (yB.max - yB.min);
        const y = area.top + area.plotH - yFrac * area.plotH;
        ctx.beginPath();
        ctx.arc(x, y, 4, 0, Math.PI * 2);
        ctx.fill();
      }
    });
  }

  private _drawPie(
    ctx: CanvasRenderingContext2D,
    w: number,
    h: number
  ): void {
    // Use first dataset for pie
    const ds = this.datasets[0];
    if (!ds) return;

    const cx = w / 2;
    const cy = h / 2;
    const radius = Math.min(w, h) / 2 - 30;
    const total = ds.data.reduce((a, b) => a + b, 0);
    if (total === 0) return;

    let startAngle = -Math.PI / 2;
    const textCol = this._textColor();

    ds.data.forEach((val, i) => {
      const sliceAngle = (val / total) * Math.PI * 2;
      const color = this._getColor(i);

      ctx.fillStyle = color;
      ctx.beginPath();
      ctx.moveTo(cx, cy);
      ctx.arc(cx, cy, radius, startAngle, startAngle + sliceAngle);
      ctx.closePath();
      ctx.fill();

      // Separator line
      ctx.strokeStyle = "rgba(0,0,0,0.2)";
      ctx.lineWidth = 1;
      ctx.stroke();

      // Label
      if (sliceAngle > 0.15) {
        const midAngle = startAngle + sliceAngle / 2;
        const labelR = radius * 0.65;
        const lx = cx + Math.cos(midAngle) * labelR;
        const ly = cy + Math.sin(midAngle) * labelR;
        const pct = ((val / total) * 100).toFixed(0);

        ctx.font = "bold 12px -apple-system, BlinkMacSystemFont, sans-serif";
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillStyle = "#fff";
        ctx.fillText(`${pct}%`, lx, ly);
      }

      startAngle += sliceAngle;
    });
  }

  render() {
    return html`
      <div class="chart-card">
        ${this.title
          ? html`
              <div class="chart-header">
                <span class="chart-title">${this.title}</span>
              </div>
            `
          : nothing}
        <div class="chart-body">
          <canvas></canvas>
        </div>
        ${this.datasets.length > 1 ||
        (this.type === "pie" && this.datasets[0]?.data.length > 1)
          ? html`
              <div class="legend">
                ${this.type === "pie"
                  ? this.labels.map(
                      (label, i) => html`
                        <span class="legend-item">
                          <span
                            class="legend-dot"
                            style="background:${this._getColor(i)}"
                          ></span>
                          ${label}
                        </span>
                      `
                    )
                  : this.datasets.map(
                      (ds, i) => html`
                        <span class="legend-item">
                          <span
                            class="legend-dot"
                            style="background:${this._datasetColor(ds, i)}"
                          ></span>
                          ${ds.label}
                        </span>
                      `
                    )}
              </div>
            `
          : nothing}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "haira-ui-chart": HairaChart;
  }
}
