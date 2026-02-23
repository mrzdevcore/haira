<script setup lang="ts">
import { ref, computed } from "vue";

interface CodeTabData { name: string; language?: string; code: string }

const props = defineProps<{
  title?: string;
  language?: string;
  code?: string;
  tabs?: CodeTabData[];
}>();

const activeTab = ref(0);
const copied = ref(false);
const activeCode = computed(() => props.tabs?.[activeTab.value]?.code ?? props.code ?? "");
const activeLang = computed(() => props.tabs?.[activeTab.value]?.language ?? props.language ?? "");

function handleCopy() {
  navigator.clipboard.writeText(activeCode.value);
  copied.value = true;
  setTimeout(() => (copied.value = false), 1500);
}
</script>

<template>
  <div class="arp-code-block" style="background: #111118; border-radius: 8px; overflow: hidden">
    <div style="display: flex; align-items: center; padding: 8px 12px; border-bottom: 1px solid #1a1a2e; gap: 8px">
      <span v-if="title" style="font-size: 13px; font-weight: 600; color: #e0e0e0">{{ title }}</span>
      <span v-if="activeLang" :style="{ fontSize: '11px', color: '#666', marginLeft: title ? '0' : 'auto' }">{{ activeLang }}</span>
      <div v-if="tabs && tabs.length > 1" style="display: flex; gap: 4px; margin-left: 8px">
        <button v-for="(tab, i) in tabs" :key="i" :style="{ background: i === activeTab ? '#2a2a3e' : 'transparent', color: i === activeTab ? '#e0e0e0' : '#666', border: 'none', borderRadius: '6px', padding: '3px 8px', fontSize: '11px', cursor: 'pointer' }" @click="activeTab = i">{{ tab.name }}</button>
      </div>
      <button :style="{ marginLeft: 'auto', background: 'none', border: 'none', color: copied ? '#22c55e' : '#666', cursor: 'pointer', fontSize: '12px' }" @click="handleCopy">{{ copied ? "Copied" : "Copy" }}</button>
    </div>
    <pre style="margin: 0; padding: 12px 16px; overflow: auto; max-height: 480px; font-size: 13px; line-height: 1.5; color: #e0e0e0"><code>{{ activeCode }}</code></pre>
  </div>
</template>
