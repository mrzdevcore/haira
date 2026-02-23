<script setup lang="ts">
import { computed } from "vue";

const STATUS_COLORS: Record<string, string> = {
  success: "#22c55e",
  error: "#ef4444",
  warning: "#f59e0b",
  info: "#3b82f6",
};

const STATUS_ICONS: Record<string, string> = {
  success: "\u2713",
  error: "\u2717",
  warning: "\u26A0",
  info: "\u2139",
};

const props = defineProps<{
  status: "success" | "error" | "warning" | "info";
  title: string;
  message?: string;
  sections?: Array<{ label: string; content: string; style?: string }>;
}>();

const color = computed(() => STATUS_COLORS[props.status] ?? STATUS_COLORS.info);
const icon = computed(() => STATUS_ICONS[props.status] ?? "\u2139");
</script>

<template>
  <div
    class="arp-status-card"
    :style="{ borderLeft: `3px solid ${color}`, background: '#111118', borderRadius: '8px', padding: '12px 16px' }"
  >
    <div style="display: flex; align-items: center; gap: 8px">
      <span :style="{ color, fontSize: '16px' }">{{ icon }}</span>
      <span style="font-weight: 600; color: #e0e0e0; font-size: 14px">{{ title }}</span>
    </div>
    <div v-if="message" style="margin-top: 8px; color: #a0a0b0; font-size: 13px; line-height: 1.5">{{ message }}</div>
    <div v-if="sections && sections.length > 0" style="margin-top: 10px; display: flex; flex-direction: column; gap: 6px">
      <div v-for="(s, i) in sections" :key="i">
        <div style="font-size: 11px; color: #666; text-transform: uppercase; letter-spacing: 0.5px">{{ s.label }}</div>
        <div :style="{ fontSize: '13px', color: s.style ? (STATUS_COLORS[s.style] ?? '#e0e0e0') : '#e0e0e0' }">{{ s.content }}</div>
      </div>
    </div>
  </div>
</template>
