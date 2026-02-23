<script setup lang="ts">
import { ref } from "vue";

const props = withDefaults(
  defineProps<{ title: string; options: string[]; style?: "buttons" | "list"; _restored?: boolean }>(),
  { style: "buttons", _restored: false },
);

const emit = defineEmits<{ input: [text: string] }>();
const selected = ref<string | null>(null);
const isDisabled = () => !!selected.value || props._restored;

function handleSelect(option: string) {
  selected.value = option;
  emit("input", option);
}
</script>

<template>
  <div class="arp-choices" style="background: #111118; border-radius: 8px; padding: 16px">
    <div style="font-weight: 600; font-size: 14px; color: #e0e0e0; margin-bottom: 10px">{{ title }}</div>
    <div v-if="isDisabled()" style="font-size: 12px; color: #666">Selected: {{ selected ?? "(from history)" }}</div>
    <div v-else-if="style === 'list'" style="display: flex; flex-direction: column; gap: 6px">
      <button v-for="(opt, i) in options" :key="i" style="display: flex; align-items: center; gap: 10px; background: #1a1a2e; color: #e0e0e0; border: 1px solid #2a2a3e; border-radius: 8px; padding: 10px 14px; cursor: pointer; font-size: 13px; text-align: left" @click="handleSelect(opt)">
        <span style="width: 16px; height: 16px; border-radius: 50%; border: 2px solid #444; flex-shrink: 0" />{{ opt }}
      </button>
    </div>
    <div v-else style="display: flex; flex-wrap: wrap; gap: 8px">
      <button v-for="(opt, i) in options" :key="i" style="background: #1a1a2e; color: #e0e0e0; border: 1px solid #2a2a3e; border-radius: 20px; padding: 8px 16px; cursor: pointer; font-size: 13px" @click="handleSelect(opt)">{{ opt }}</button>
    </div>
  </div>
</template>
