<script setup lang="ts">
import { reactive } from "vue";

interface FormField { name: string; label?: string; field_type?: string; value?: string; required?: boolean; options?: string[] }

const props = withDefaults(
  defineProps<{ title?: string; fields: FormField[]; submit_label?: string; submit_action?: string }>(),
  { title: undefined, submit_label: "Submit", submit_action: undefined },
);

const emit = defineEmits<{ input: [text: string] }>();
const values = reactive<Record<string, string>>({});
for (const f of props.fields) values[f.name] = f.value ?? "";

function handleSubmit() {
  const payload = props.submit_action ? `[Form: ${props.submit_action}] ${JSON.stringify(values)}` : JSON.stringify(values);
  emit("input", payload);
}
</script>

<template>
  <div class="arp-form" style="background: #111118; border-radius: 8px; padding: 16px">
    <div v-if="title" style="font-weight: 600; font-size: 14px; color: #e0e0e0; margin-bottom: 12px">{{ title }}</div>
    <div style="display: flex; flex-direction: column; gap: 12px">
      <div v-for="field in fields" :key="field.name">
        <label style="display: block; font-size: 12px; color: #888; margin-bottom: 4px">
          {{ field.label ?? field.name }}<span v-if="field.required" style="color: #ef4444"> *</span>
        </label>
        <select v-if="field.options" :value="values[field.name]" style="width: 100%; padding: 8px 10px; background: #1a1a2e; color: #e0e0e0; border: 1px solid #2a2a3e; border-radius: 6px; font-size: 13px" @input="(e) => { values[field.name] = (e.target as HTMLSelectElement).value }">
          <option value="">Select...</option>
          <option v-for="opt in field.options" :key="opt" :value="opt">{{ opt }}</option>
        </select>
        <input v-else :type="field.field_type === 'number' ? 'number' : 'text'" :value="values[field.name]" style="width: 100%; padding: 8px 10px; background: #1a1a2e; color: #e0e0e0; border: 1px solid #2a2a3e; border-radius: 6px; font-size: 13px; box-sizing: border-box" @input="(e) => { values[field.name] = (e.target as HTMLInputElement).value }" />
      </div>
    </div>
    <button style="margin-top: 16px; background: #e8a317; color: #000; border: none; border-radius: 8px; padding: 8px 20px; font-weight: 600; font-size: 13px; cursor: pointer" @click="handleSubmit">{{ submit_label }}</button>
  </div>
</template>
