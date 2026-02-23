<script setup lang="ts">
import { ref } from "vue";

const props = withDefaults(
  defineProps<{ title: string; message?: string; confirm_label?: string; deny_label?: string; _restored?: boolean }>(),
  { message: undefined, confirm_label: "Confirm", deny_label: "Deny", _restored: false },
);

const emit = defineEmits<{ input: [text: string] }>();
const chosen = ref<string | null>(null);
const disabled = () => !!chosen.value || props._restored;

function handleChoice(label: string, action: string) {
  chosen.value = label;
  emit("input", `[${action}] ${label}`);
}
</script>

<template>
  <div class="arp-confirm" style="background: #111118; border-radius: 8px; padding: 16px">
    <div style="font-weight: 600; font-size: 14px; color: #e0e0e0">{{ title }}</div>
    <div v-if="message" style="margin-top: 6px; font-size: 13px; color: #a0a0b0; line-height: 1.5">{{ message }}</div>
    <div v-if="disabled()" style="margin-top: 12px; font-size: 12px; color: #666">Selection made</div>
    <div v-else style="margin-top: 12px; display: flex; gap: 8px">
      <button style="background: #1a1a2e; color: #a0a0b0; border: 1px solid #2a2a3e; border-radius: 8px; padding: 8px 16px; cursor: pointer; font-size: 13px" @click="handleChoice(deny_label, 'Denied')">{{ deny_label }}</button>
      <button style="background: #e8a317; color: #000; border: none; border-radius: 8px; padding: 8px 16px; cursor: pointer; font-weight: 600; font-size: 13px" @click="handleChoice(confirm_label, 'Confirmed')">{{ confirm_label }}</button>
    </div>
  </div>
</template>
