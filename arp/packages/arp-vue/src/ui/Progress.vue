<script setup lang="ts">
const STATUS: Record<string, { color: string; icon: string }> = {
  done: { color: "#22c55e", icon: "\u2713" }, completed: { color: "#22c55e", icon: "\u2713" },
  running: { color: "#e8a317", icon: "\u25CF" }, active: { color: "#e8a317", icon: "\u25CF" },
  failed: { color: "#ef4444", icon: "\u2717" }, error: { color: "#ef4444", icon: "\u2717" },
  pending: { color: "#555", icon: "\u25CB" },
};

defineProps<{
  title?: string;
  steps: Array<{ name: string; status: string; detail?: string }>;
}>();

function getStatus(status: string) { return STATUS[status] ?? STATUS.pending; }
</script>

<template>
  <div class="arp-progress" style="background: #111118; border-radius: 8px; padding: 12px 16px">
    <div v-if="title" style="font-weight: 600; font-size: 14px; color: #e0e0e0; margin-bottom: 12px">{{ title }}</div>
    <div style="display: flex; flex-direction: column">
      <div v-for="(step, i) in steps" :key="i" style="display: flex; gap: 12px; align-items: flex-start">
        <div style="display: flex; flex-direction: column; align-items: center; width: 20px">
          <span :style="{ color: getStatus(step.status).color, fontSize: '14px', lineHeight: '24px' }">{{ getStatus(step.status).icon }}</span>
          <div v-if="i < steps.length - 1" style="width: 1px; height: 20px; background: #2a2a3e" />
        </div>
        <div :style="{ paddingBottom: i < steps.length - 1 ? '8px' : '0' }">
          <div style="font-size: 13px; color: #e0e0e0">{{ step.name }}</div>
          <div v-if="step.detail" style="font-size: 12px; color: #666; margin-top: 2px">{{ step.detail }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
