import { BaseComponent, animateInCSS, cardCSS } from "../core";
import type { ChartProps, ChartDataset } from "../core/types";

const DEFAULT_COLORS = [
  "#e8a317", "#3b82f6", "#22c55e", "#ef4444", "#a855f7",
  "#f59e0b", "#06b6d4", "#ec4899", "#84cc16", "#f97316",
];

export class HairaChart extends BaseComponent<ChartProps> {
  private canvas!: HTMLCanvasElement;
  private ctx!: CanvasRenderingContext2D;

  protected render() {
    return `
      <div class="card">
        <div class="header" id="header" style="display:none">
          <span class="title" id="title"></span>
        </div>
        <div class="canvas-wrap">
          <canvas id="canvas"></canvas>
        </div>
        <div class="legend" id="legend"></div>
      </div>`;
  }

  protected styles() {
    return `
      ${animateInCSS}
      .card { ${cardCSS} }
      .header {
        padding: 0.5rem 0.75rem; border-bottom: 1px solid var(--haira-border);
        font-size: 0.78rem; font-weight: 600; color: var(--haira-text);
      }
      .canvas-wrap { padding: 0.75rem; }
      canvas { width: 100%; display: block; }
      .legend {
        display: flex; flex-wrap: wrap; gap: 0.75rem; padding: 0 0.75rem 0.6rem;
        font-size: 0.7rem; color: var(--haira-text-dim);
      }
      .legend-item { display: flex; align-items: center; gap: 0.3rem; }
      .legend-dot { width: 8px; height: 8px; border-radius: 2px; flex-shrink: 0; }`;
  }

  protected onMount() {
    this.canvas = this.$("canvas") as HTMLCanvasElement;
    this.ctx = this.canvas.getContext("2d")!;
  }

  protected onUpdate() {
    const { title, type, labels = [], datasets = [], height = 240 } = this.props;

    const header = this.$("header");
    if (title) {
      this.$("title").textContent = title;
      header.style.display = "";
    }

    // Set canvas size (2x for retina)
    const dpr = window.devicePixelRatio || 1;
    const width = this.canvas.parentElement!.clientWidth;
    this.canvas.width = width * dpr;
    this.canvas.height = height * dpr;
    this.canvas.style.height = `${height}px`;
    this.ctx.scale(dpr, dpr);

    // Assign colors
    const coloredDatasets = datasets.map((ds, i) => ({
      ...ds,
      color: ds.color || DEFAULT_COLORS[i % DEFAULT_COLORS.length],
    }));

    switch (type) {
      case "bar":
        this.drawBar(labels, coloredDatasets, width, height);
        break;
      case "pie":
        this.drawPie(coloredDatasets, width, height);
        break;
      case "line":
      case "area":
        this.drawLine(labels, coloredDatasets, width, height, type === "area");
        break;
      case "scatter":
        this.drawScatter(labels, coloredDatasets, width, height);
        break;
      default:
        this.drawBar(labels, coloredDatasets, width, height);
    }

    // Legend
    const legend = this.$("legend");
    if (coloredDatasets.length > 1 || type === "pie") {
      const items = type === "pie" ? labels : coloredDatasets.map((ds) => ds.label);
      const colors = type === "pie"
        ? labels.map((_, i) => DEFAULT_COLORS[i % DEFAULT_COLORS.length])
        : coloredDatasets.map((ds) => ds.color!);
      legend.innerHTML = items
        .map(
          (name, i) =>
            `<span class="legend-item"><span class="legend-dot" style="background:${colors[i]}"></span>${this.esc(name)}</span>`,
        )
        .join("");
    } else {
      legend.innerHTML = "";
    }
  }

  private drawBar(labels: string[], datasets: (ChartDataset & { color: string })[], w: number, h: number) {
    const ctx = this.ctx;
    const pad = { top: 10, right: 10, bottom: 30, left: 45 };
    const cw = w - pad.left - pad.right;
    const ch = h - pad.top - pad.bottom;

    const allValues = datasets.flatMap((ds) => ds.data);
    const max = Math.max(...allValues, 0) * 1.1 || 1;

    // Grid
    ctx.strokeStyle = "rgba(63, 63, 70, 0.3)";
    ctx.lineWidth = 0.5;
    for (let i = 0; i <= 4; i++) {
      const y = pad.top + ch - (i / 4) * ch;
      ctx.beginPath();
      ctx.moveTo(pad.left, y);
      ctx.lineTo(w - pad.right, y);
      ctx.stroke();

      ctx.fillStyle = "#71717a";
      ctx.font = "10px -apple-system, sans-serif";
      ctx.textAlign = "right";
      ctx.fillText(this.formatNum((max / 4) * i), pad.left - 5, y + 3);
    }

    // Bars
    const groupCount = labels.length;
    const dsCount = datasets.length;
    const groupWidth = cw / groupCount;
    const barWidth = Math.min(groupWidth * 0.7 / dsCount, 40);
    const totalBarWidth = barWidth * dsCount;

    for (let g = 0; g < groupCount; g++) {
      const groupX = pad.left + g * groupWidth + (groupWidth - totalBarWidth) / 2;

      for (let d = 0; d < dsCount; d++) {
        const val = datasets[d].data[g] || 0;
        const barH = (val / max) * ch;
        const x = groupX + d * barWidth;
        const y = pad.top + ch - barH;

        ctx.fillStyle = datasets[d].color;
        ctx.beginPath();
        ctx.roundRect(x, y, barWidth - 1, barH, [3, 3, 0, 0]);
        ctx.fill();
      }

      // Label
      ctx.fillStyle = "#71717a";
      ctx.font = "10px -apple-system, sans-serif";
      ctx.textAlign = "center";
      const labelX = pad.left + g * groupWidth + groupWidth / 2;
      ctx.fillText(this.truncLabel(labels[g], 10), labelX, h - 8);
    }
  }

  private drawLine(labels: string[], datasets: (ChartDataset & { color: string })[], w: number, h: number, fill: boolean) {
    const ctx = this.ctx;
    const pad = { top: 10, right: 10, bottom: 30, left: 45 };
    const cw = w - pad.left - pad.right;
    const ch = h - pad.top - pad.bottom;

    const allValues = datasets.flatMap((ds) => ds.data);
    const max = Math.max(...allValues, 0) * 1.1 || 1;

    // Grid
    ctx.strokeStyle = "rgba(63, 63, 70, 0.3)";
    ctx.lineWidth = 0.5;
    for (let i = 0; i <= 4; i++) {
      const y = pad.top + ch - (i / 4) * ch;
      ctx.beginPath();
      ctx.moveTo(pad.left, y);
      ctx.lineTo(w - pad.right, y);
      ctx.stroke();

      ctx.fillStyle = "#71717a";
      ctx.font = "10px -apple-system, sans-serif";
      ctx.textAlign = "right";
      ctx.fillText(this.formatNum((max / 4) * i), pad.left - 5, y + 3);
    }

    // X labels
    for (let i = 0; i < labels.length; i++) {
      const x = pad.left + (i / (labels.length - 1 || 1)) * cw;
      ctx.fillStyle = "#71717a";
      ctx.font = "10px -apple-system, sans-serif";
      ctx.textAlign = "center";
      ctx.fillText(this.truncLabel(labels[i], 10), x, h - 8);
    }

    // Lines
    for (const ds of datasets) {
      ctx.strokeStyle = ds.color;
      ctx.lineWidth = 2;
      ctx.lineJoin = "round";
      ctx.beginPath();

      const points: [number, number][] = [];
      for (let i = 0; i < ds.data.length; i++) {
        const x = pad.left + (i / (ds.data.length - 1 || 1)) * cw;
        const y = pad.top + ch - (ds.data[i] / max) * ch;
        points.push([x, y]);
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      }
      ctx.stroke();

      if (fill && points.length > 0) {
        ctx.globalAlpha = 0.1;
        ctx.fillStyle = ds.color;
        ctx.beginPath();
        ctx.moveTo(points[0][0], pad.top + ch);
        for (const [x, y] of points) ctx.lineTo(x, y);
        ctx.lineTo(points[points.length - 1][0], pad.top + ch);
        ctx.closePath();
        ctx.fill();
        ctx.globalAlpha = 1;
      }

      // Dots
      for (const [x, y] of points) {
        ctx.fillStyle = ds.color;
        ctx.beginPath();
        ctx.arc(x, y, 3, 0, Math.PI * 2);
        ctx.fill();
      }
    }
  }

  private drawPie(datasets: (ChartDataset & { color: string })[], w: number, h: number) {
    const ctx = this.ctx;
    const data = datasets[0]?.data || [];
    const total = data.reduce((a, b) => a + b, 0) || 1;
    const cx = w / 2;
    const cy = h / 2;
    const radius = Math.min(cx, cy) - 20;

    let startAngle = -Math.PI / 2;
    for (let i = 0; i < data.length; i++) {
      const sliceAngle = (data[i] / total) * Math.PI * 2;
      ctx.fillStyle = DEFAULT_COLORS[i % DEFAULT_COLORS.length];
      ctx.beginPath();
      ctx.moveTo(cx, cy);
      ctx.arc(cx, cy, radius, startAngle, startAngle + sliceAngle);
      ctx.closePath();
      ctx.fill();

      // Separator
      ctx.strokeStyle = "var(--haira-bg-card, #0f0f12)";
      ctx.lineWidth = 2;
      ctx.stroke();

      startAngle += sliceAngle;
    }
  }

  private drawScatter(labels: string[], datasets: (ChartDataset & { color: string })[], w: number, h: number) {
    // For scatter, treat data as pairs [x,y] alternating in the array
    // or simply plot index vs value (simplified scatter)
    this.drawLine(labels, datasets, w, h, false);
  }

  private formatNum(n: number): string {
    if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
    if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
    return n % 1 === 0 ? String(n) : n.toFixed(1);
  }

  private truncLabel(s: string, max: number): string {
    return s.length > max ? s.slice(0, max - 1) + "\u2026" : s;
  }
}
