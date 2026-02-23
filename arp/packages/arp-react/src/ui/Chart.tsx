import { useRef, useEffect } from "react";
import type { ChartProps } from "@haira/arp";

const DEFAULT_COLORS = [
  "#e8a317", "#3b82f6", "#22c55e", "#ef4444", "#a78bfa",
  "#f59e0b", "#06b6d4", "#ec4899", "#8b5cf6", "#14b8a6",
];

export function Chart({
  type,
  title,
  labels,
  datasets,
  height = 300,
}: ChartProps & { onInput?: (text: string) => void }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    const w = canvas.clientWidth;
    const h = canvas.clientHeight;
    canvas.width = w * dpr;
    canvas.height = h * dpr;
    ctx.scale(dpr, dpr);
    ctx.clearRect(0, 0, w, h);

    if (type === "pie") {
      drawPie(ctx, w, h, labels, datasets);
    } else {
      drawCartesian(ctx, w, h, type, labels, datasets);
    }
  }, [type, labels, datasets]);

  return (
    <div className="arp-chart" style={{ background: "#111118", borderRadius: 8, overflow: "hidden" }}>
      {title && (
        <div style={{ padding: "10px 16px", fontWeight: 600, fontSize: 14, color: "#e0e0e0" }}>
          {title}
        </div>
      )}
      <div style={{ padding: "8px 16px 16px" }}>
        <canvas
          ref={canvasRef}
          style={{ width: "100%", height }}
        />
      </div>
      {/* Legend */}
      {(datasets.length > 1 || type === "pie") && (
        <div style={{ display: "flex", flexWrap: "wrap", gap: 12, padding: "0 16px 12px", justifyContent: "center" }}>
          {(type === "pie" ? labels : datasets.map((d) => d.label)).map((label, i) => (
            <div key={i} style={{ display: "flex", alignItems: "center", gap: 4, fontSize: 12, color: "#a0a0b0" }}>
              <span style={{ width: 10, height: 10, borderRadius: 2, background: datasets[i]?.color ?? DEFAULT_COLORS[i % DEFAULT_COLORS.length] }} />
              {label}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Canvas drawing helpers
// ---------------------------------------------------------------------------

function drawPie(
  ctx: CanvasRenderingContext2D,
  w: number,
  h: number,
  labels: string[],
  datasets: ChartProps["datasets"],
) {
  const data = datasets[0]?.data ?? [];
  const total = data.reduce((a, b) => a + b, 0);
  if (total === 0) return;

  const cx = w / 2;
  const cy = h / 2;
  const r = Math.min(cx, cy) - 20;
  let angle = -Math.PI / 2;

  data.forEach((v, i) => {
    const sliceAngle = (v / total) * Math.PI * 2;
    ctx.beginPath();
    ctx.moveTo(cx, cy);
    ctx.arc(cx, cy, r, angle, angle + sliceAngle);
    ctx.closePath();
    ctx.fillStyle = datasets[0]?.color
      ? datasets[0].color
      : DEFAULT_COLORS[i % DEFAULT_COLORS.length];
    ctx.fill();
    angle += sliceAngle;
  });
}

function drawCartesian(
  ctx: CanvasRenderingContext2D,
  w: number,
  h: number,
  type: string,
  labels: string[],
  datasets: ChartProps["datasets"],
) {
  const pad = { top: 10, right: 20, bottom: 30, left: 50 };
  const plotW = w - pad.left - pad.right;
  const plotH = h - pad.top - pad.bottom;

  // Find data range
  let maxVal = 0;
  for (const ds of datasets) {
    for (const v of ds.data) {
      if (v > maxVal) maxVal = v;
    }
  }
  if (maxVal === 0) maxVal = 1;

  // Draw axes
  ctx.strokeStyle = "#2a2a3e";
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(pad.left, pad.top);
  ctx.lineTo(pad.left, pad.top + plotH);
  ctx.lineTo(pad.left + plotW, pad.top + plotH);
  ctx.stroke();

  // X labels
  ctx.fillStyle = "#666";
  ctx.font = "11px sans-serif";
  ctx.textAlign = "center";
  const step = plotW / Math.max(labels.length - 1, 1);
  labels.forEach((label, i) => {
    const x = pad.left + i * step;
    ctx.fillText(label, x, h - 8);
  });

  // Draw datasets
  datasets.forEach((ds, di) => {
    const color = ds.color ?? DEFAULT_COLORS[di % DEFAULT_COLORS.length];
    ctx.strokeStyle = color;
    ctx.fillStyle = color;
    ctx.lineWidth = 2;

    const points = ds.data.map((v, i) => ({
      x: pad.left + i * step,
      y: pad.top + plotH - (v / maxVal) * plotH,
    }));

    if (type === "bar" || type === "scatter") {
      const barW = step / (datasets.length + 1);
      points.forEach((p, i) => {
        if (type === "bar") {
          const x = p.x - (datasets.length * barW) / 2 + di * barW;
          const barH = pad.top + plotH - p.y;
          ctx.globalAlpha = 0.8;
          ctx.fillRect(x, p.y, barW - 2, barH);
          ctx.globalAlpha = 1;
        } else {
          ctx.beginPath();
          ctx.arc(p.x, p.y, 4, 0, Math.PI * 2);
          ctx.fill();
        }
      });
    } else {
      // Line / area
      if (points.length > 0) {
        if (type === "area") {
          ctx.beginPath();
          ctx.moveTo(points[0].x, pad.top + plotH);
          points.forEach((p) => ctx.lineTo(p.x, p.y));
          ctx.lineTo(points[points.length - 1].x, pad.top + plotH);
          ctx.closePath();
          ctx.globalAlpha = 0.15;
          ctx.fill();
          ctx.globalAlpha = 1;
        }
        ctx.beginPath();
        points.forEach((p, i) => (i === 0 ? ctx.moveTo(p.x, p.y) : ctx.lineTo(p.x, p.y)));
        ctx.stroke();
        // Dots
        points.forEach((p) => {
          ctx.beginPath();
          ctx.arc(p.x, p.y, 3, 0, Math.PI * 2);
          ctx.fill();
        });
      }
    }
  });
}
