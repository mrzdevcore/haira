<script setup lang="ts">
import { ref, onMounted, watch } from "vue";

const COLORS = ["#e8a317", "#3b82f6", "#22c55e", "#ef4444", "#a78bfa", "#f59e0b", "#06b6d4", "#ec4899", "#8b5cf6", "#14b8a6"];

interface ChartDatasetLocal { label: string; data: number[]; color?: string }

const props = withDefaults(
  defineProps<{
    type: "line" | "bar" | "pie" | "scatter" | "area";
    title?: string;
    labels: string[];
    datasets: ChartDatasetLocal[];
    height?: number;
  }>(),
  { title: undefined, height: 300 },
);

const canvasRef = ref<HTMLCanvasElement | null>(null);

function draw() {
  const canvas = canvasRef.value;
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  const dpr = window.devicePixelRatio || 1;
  const w = canvas.clientWidth, hgt = canvas.clientHeight;
  canvas.width = w * dpr; canvas.height = hgt * dpr;
  ctx.scale(dpr, dpr); ctx.clearRect(0, 0, w, hgt);

  if (props.type === "pie") {
    const data = props.datasets[0]?.data ?? [];
    const total = data.reduce((a, b) => a + b, 0);
    if (total === 0) return;
    const cx = w / 2, cy = hgt / 2, r = Math.min(cx, cy) - 20;
    let angle = -Math.PI / 2;
    data.forEach((v, i) => {
      const sa = (v / total) * Math.PI * 2;
      ctx.beginPath(); ctx.moveTo(cx, cy); ctx.arc(cx, cy, r, angle, angle + sa); ctx.closePath();
      ctx.fillStyle = COLORS[i % COLORS.length]; ctx.fill(); angle += sa;
    });
  } else {
    const pad = { top: 10, right: 20, bottom: 30, left: 50 };
    const pw = w - pad.left - pad.right, ph = hgt - pad.top - pad.bottom;
    let maxVal = 0;
    for (const ds of props.datasets) for (const v of ds.data) if (v > maxVal) maxVal = v;
    if (maxVal === 0) maxVal = 1;
    ctx.strokeStyle = "#2a2a3e"; ctx.lineWidth = 1;
    ctx.beginPath(); ctx.moveTo(pad.left, pad.top); ctx.lineTo(pad.left, pad.top + ph); ctx.lineTo(pad.left + pw, pad.top + ph); ctx.stroke();
    ctx.fillStyle = "#666"; ctx.font = "11px sans-serif"; ctx.textAlign = "center";
    const step = pw / Math.max(props.labels.length - 1, 1);
    props.labels.forEach((l, i) => ctx.fillText(l, pad.left + i * step, hgt - 8));
    props.datasets.forEach((ds, di) => {
      const col = ds.color ?? COLORS[di % COLORS.length];
      ctx.strokeStyle = col; ctx.fillStyle = col; ctx.lineWidth = 2;
      const pts = ds.data.map((v, i) => ({ x: pad.left + i * step, y: pad.top + ph - (v / maxVal) * ph }));
      if (props.type === "bar") {
        const bw = step / (props.datasets.length + 1);
        pts.forEach((p) => { ctx.globalAlpha = 0.8; ctx.fillRect(p.x - (props.datasets.length * bw) / 2 + di * bw, p.y, bw - 2, pad.top + ph - p.y); ctx.globalAlpha = 1; });
      } else if (props.type === "scatter") {
        pts.forEach((p) => { ctx.beginPath(); ctx.arc(p.x, p.y, 4, 0, Math.PI * 2); ctx.fill(); });
      } else if (pts.length > 0) {
        if (props.type === "area") { ctx.beginPath(); ctx.moveTo(pts[0].x, pad.top + ph); pts.forEach((p) => ctx.lineTo(p.x, p.y)); ctx.lineTo(pts[pts.length - 1].x, pad.top + ph); ctx.closePath(); ctx.globalAlpha = 0.15; ctx.fill(); ctx.globalAlpha = 1; }
        ctx.beginPath(); pts.forEach((p, i) => i === 0 ? ctx.moveTo(p.x, p.y) : ctx.lineTo(p.x, p.y)); ctx.stroke();
        pts.forEach((p) => { ctx.beginPath(); ctx.arc(p.x, p.y, 3, 0, Math.PI * 2); ctx.fill(); });
      }
    });
  }
}

onMounted(draw);
watch(() => [props.type, props.labels, props.datasets], draw, { deep: true });
function legendLabels() { return props.type === "pie" ? props.labels : props.datasets.map((d) => d.label); }
function legendColor(i: number) { return props.datasets[i]?.color ?? COLORS[i % COLORS.length]; }
function showLegend() { return props.datasets.length > 1 || props.type === "pie"; }
</script>

<template>
  <div class="arp-chart" style="background: #111118; border-radius: 8px; overflow: hidden">
    <div v-if="title" style="padding: 10px 16px; font-weight: 600; font-size: 14px; color: #e0e0e0">{{ title }}</div>
    <div style="padding: 8px 16px 16px">
      <canvas ref="canvasRef" :style="{ width: '100%', height: `${height}px` }" />
    </div>
    <div v-if="showLegend()" style="display: flex; flex-wrap: wrap; gap: 12px; padding: 0 16px 12px; justify-content: center">
      <div v-for="(label, i) in legendLabels()" :key="i" style="display: flex; align-items: center; gap: 4px; font-size: 12px; color: #a0a0b0">
        <span :style="{ width: '10px', height: '10px', borderRadius: '2px', background: legendColor(i) }" />
        {{ label }}
      </div>
    </div>
  </div>
</template>
