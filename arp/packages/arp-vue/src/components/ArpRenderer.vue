<script setup lang="ts">
import { inject, computed } from "vue";
import { ArpComponentsKey } from "../provide-keys";

interface ToolRenderEventLocal {
  component: string;
  props: Record<string, any>;
}

const props = defineProps<{
  event: ToolRenderEventLocal;
  components?: Record<string, any>;
  restored?: boolean;
}>();

const emit = defineEmits<{
  input: [text: string];
}>();

const contextComponents = inject(ArpComponentsKey, {});

const registry = computed(() => ({
  ...contextComponents,
  ...props.components,
}));

const Component = computed(() => registry.value[props.event.component]);
</script>

<template>
  <component
    v-if="Component"
    :is="Component"
    v-bind="event.props"
    :_restored="restored"
    @input="(text: string) => emit('input', text)"
  />
  <pre
    v-else
    style="background: #1a1a2e; color: #a0a0b0; padding: 12px; border-radius: 8px; font-size: 13px; overflow: auto"
  >{{ JSON.stringify({ component: event.component, props: event.props }, null, 2) }}</pre>
</template>
