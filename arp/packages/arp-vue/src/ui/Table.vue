<script setup lang="ts">
import { ref, computed } from "vue";

interface TabData { name: string; headers: string[]; rows: string[][]; highlight?: number[] }

const props = defineProps<{
  title?: string;
  headers: string[];
  rows: string[][];
  tabs?: TabData[];
  highlight?: number[];
}>();

const activeTab = ref(0);
const activeHeaders = computed(() => props.tabs?.[activeTab.value]?.headers ?? props.headers);
const activeRows = computed(() => props.tabs?.[activeTab.value]?.rows ?? props.rows);
const activeHighlight = computed(() => props.tabs?.[activeTab.value]?.highlight ?? props.highlight);
</script>

<template>
  <div class="arp-table" style="background: #111118; border-radius: 8px; overflow: hidden">
    <div v-if="title || tabs" style="padding: 10px 16px; border-bottom: 1px solid #1a1a2e; display: flex; align-items: center; gap: 12px">
      <span v-if="title" style="font-weight: 600; font-size: 14px; color: #e0e0e0">{{ title }}</span>
      <div v-if="tabs && tabs.length > 1" style="display: flex; gap: 4px; margin-left: auto">
        <button v-for="(tab, i) in tabs" :key="i" :style="{ background: i === activeTab ? '#2a2a3e' : 'transparent', color: i === activeTab ? '#e0e0e0' : '#666', border: 'none', borderRadius: '6px', padding: '4px 10px', fontSize: '12px', cursor: 'pointer' }" @click="activeTab = i">{{ tab.name }}</button>
      </div>
    </div>
    <div style="overflow-x: auto">
      <table style="width: 100%; border-collapse: collapse; font-size: 13px">
        <thead>
          <tr>
            <th v-for="(h, i) in activeHeaders" :key="i" style="padding: 8px 12px; text-align: left; color: #888; font-weight: 500; border-bottom: 1px solid #1a1a2e; white-space: nowrap">{{ h }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, ri) in activeRows" :key="ri" :style="{ background: activeHighlight?.includes(ri) ? 'rgba(232,163,23,0.08)' : undefined }">
            <td v-for="(cell, ci) in row" :key="ci" style="padding: 8px 12px; color: #e0e0e0; border-bottom: 1px solid #0a0a12">{{ cell }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
