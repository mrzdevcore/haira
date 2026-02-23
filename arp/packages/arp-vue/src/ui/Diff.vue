<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string;
    before_label?: string;
    after_label?: string;
    before: string;
    after: string;
    language?: string;
  }>(),
  { title: undefined, before_label: "Before", after_label: "After", language: undefined },
);
</script>

<template>
  <div class="arp-diff" style="background: #111118; border-radius: 8px; overflow: hidden">
    <div v-if="title" style="padding: 10px 16px; border-bottom: 1px solid #1a1a2e; font-weight: 600; font-size: 14px; color: #e0e0e0">{{ title }}</div>
    <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 0">
      <div style="border-right: 1px solid #1a1a2e">
        <div style="padding: 6px 12px; font-size: 11px; color: #ef4444; font-weight: 600; border-bottom: 1px solid #1a1a2e">{{ before_label }}</div>
        <pre style="margin: 0; padding: 12px 16px; overflow: auto; font-size: 13px; line-height: 1.5; color: #f87171; background: rgba(239, 68, 68, 0.04)">{{ before }}</pre>
      </div>
      <div>
        <div style="padding: 6px 12px; font-size: 11px; color: #22c55e; font-weight: 600; border-bottom: 1px solid #1a1a2e">{{ after_label }}</div>
        <pre style="margin: 0; padding: 12px 16px; overflow: auto; font-size: 13px; line-height: 1.5; color: #4ade80; background: rgba(34, 197, 94, 0.04)">{{ after }}</pre>
      </div>
    </div>
  </div>
</template>
