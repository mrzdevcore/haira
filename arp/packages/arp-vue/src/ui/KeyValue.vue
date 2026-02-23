<script setup lang="ts">
const STYLE_COLORS: Record<string, string> = { success: "#22c55e", error: "#ef4444", warning: "#f59e0b", info: "#3b82f6", code: "#a78bfa" };

defineProps<{
  title?: string;
  items: Array<{ key: string; value: string; style?: string }>;
}>();
</script>

<template>
  <div class="arp-key-value" style="background: #111118; border-radius: 8px; overflow: hidden">
    <div v-if="title" style="padding: 10px 16px; border-bottom: 1px solid #1a1a2e; font-weight: 600; font-size: 14px; color: #e0e0e0">{{ title }}</div>
    <div style="padding: 4px 0">
      <div v-for="(item, i) in items" :key="i" :style="{ display: 'flex', padding: '8px 16px', gap: '16px', fontSize: '13px', borderBottom: i < items.length - 1 ? '1px solid #0a0a12' : undefined }">
        <span style="color: #888; min-width: 120px; flex-shrink: 0">{{ item.key }}</span>
        <span :style="{ color: item.style ? (STYLE_COLORS[item.style] ?? '#e0e0e0') : '#e0e0e0', fontFamily: item.style === 'code' ? 'monospace' : 'inherit' }">{{ item.value }}</span>
      </div>
    </div>
  </div>
</template>
