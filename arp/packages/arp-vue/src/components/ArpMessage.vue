<script setup lang="ts">
import { computed } from "vue";
import ArpRenderer from "./ArpRenderer.vue";

interface ToolRenderEventLocal {
  component: string;
  props: Record<string, any>;
}

const props = withDefaults(
  defineProps<{
    role: "user" | "assistant";
    content: string;
    avatar?: string;
    file?: string;
    uiEvents?: ToolRenderEventLocal[];
    restored?: boolean;
  }>(),
  {
    avatar: undefined,
    file: undefined,
    uiEvents: undefined,
    restored: false,
  },
);

const emit = defineEmits<{
  input: [text: string];
}>();

const isUser = computed(() => props.role === "user");

function renderMarkdown(content: string): string {
  if (!content) return "";
  return content
    .replace(
      /```(\w*)\n([\s\S]*?)```/g,
      '<pre style="background:#1a1a2e;color:#e0e0e0;padding:12px;border-radius:8px;overflow:auto;font-size:13px"><code>$2</code></pre>',
    )
    .replace(
      /`([^`]+)`/g,
      '<code style="background:#2a2a3e;padding:2px 6px;border-radius:4px;font-size:13px">$1</code>',
    )
    .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
    .replace(/\*(.+?)\*/g, "<em>$1</em>")
    .replace(
      /\[([^\]]+)\]\(([^)]+)\)/g,
      '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>',
    )
    .replace(/\n\n/g, "</p><p>")
    .replace(/\n/g, "<br/>");
}

const renderedHtml = computed(() => `<p>${renderMarkdown(props.content)}</p>`);
</script>

<template>
  <div
    :class="['arp-message', `arp-message--${role}`]"
    :style="{
      display: 'flex',
      gap: '12px',
      padding: '16px 0',
      flexDirection: isUser ? 'row-reverse' : 'row',
    }"
  >
    <!-- Avatar -->
    <div
      v-if="!isUser"
      class="arp-message-avatar"
      :style="{
        width: '32px',
        height: '32px',
        borderRadius: '50%',
        background: '#2a2a3e',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        flexShrink: 0,
        fontSize: '14px',
        color: '#a0a0b0',
      }"
    >
      {{ avatar || "A" }}
    </div>

    <!-- Content -->
    <div style="max-width: 80%; min-width: 0">
      <!-- File badge -->
      <div
        v-if="file"
        class="arp-message-file"
        style="font-size: 12px; color: #e8a317; margin-bottom: 4px"
      >
        {{ file }}
      </div>

      <!-- Text content -->
      <div
        v-if="content"
        class="arp-message-bubble"
        :style="{
          background: isUser ? '#2a2a3e' : 'transparent',
          padding: isUser ? '10px 14px' : '0',
          borderRadius: isUser ? '12px' : '0',
          color: '#e0e0e0',
          fontSize: '14px',
          lineHeight: 1.6,
        }"
      >
        <!-- eslint-disable vue/no-v-html -->
        <div class="arp-message-content" v-html="renderedHtml" />
      </div>

      <!-- Generative UI components -->
      <div v-for="(event, i) in uiEvents" :key="i" style="margin-top: 8px">
        <ArpRenderer
          :event="event"
          :restored="restored"
          @input="(text: string) => emit('input', text)"
        />
      </div>
    </div>
  </div>
</template>
