<script setup lang="ts">
import { ref, watch, nextTick, provide, computed } from "vue";
import { useArpChat, type UseArpChatOptions } from "../composables/use-arp-chat";
import { ArpComponentsKey, ArpConnectedKey, ArpCapabilitiesKey } from "../provide-keys";
import ArpMessage from "./ArpMessage.vue";
import ArpActivityPanel from "./ArpActivityPanel.vue";

export interface ArpChatProps extends UseArpChatOptions {
  title?: string;
  description?: string;
  suggestions?: string[];
  avatar?: string;
  logo?: string;
  theme?: "dark" | "light";
  accentColor?: string;
  components?: Record<string, any>;
  showActivityPanel?: boolean;
}

const props = withDefaults(defineProps<ArpChatProps>(), {
  title: "Chat",
  description: undefined,
  suggestions: undefined,
  avatar: undefined,
  logo: undefined,
  theme: "dark",
  accentColor: "#e8a317",
  components: () => ({}),
  showActivityPanel: true,
});

const emit = defineEmits<{
  send: [message: string];
  "session-change": [sessionId: string];
}>();

const chat = useArpChat(props);
const input = ref("");
const panelOpen = ref(false);
const messagesEnd = ref<HTMLElement | null>(null);
const inputEl = ref<HTMLTextAreaElement | null>(null);

const isDark = computed(() => props.theme === "dark");
const showWelcome = computed(() => chat.messages.value.length === 0);

// Provide context
provide(ArpComponentsKey, props.components);
provide(ArpConnectedKey, chat.isConnected.value);
provide(ArpCapabilitiesKey, chat.capabilities.value);

// Auto-scroll
watch(
  () => chat.messages.value,
  async () => {
    await nextTick();
    messagesEnd.value?.scrollIntoView({ behavior: "smooth" });
  },
  { deep: true },
);

// Open panel on tool activity
watch(
  () => chat.runningToolCount.value,
  (count) => {
    if (count > 0 && props.showActivityPanel) panelOpen.value = true;
  },
);

// Close panel when streaming done
watch(
  () => chat.isStreaming.value,
  (streaming) => {
    if (!streaming && panelOpen.value) {
      setTimeout(() => (panelOpen.value = false), 500);
    }
  },
);

// Notify parent of session changes
watch(
  () => chat.sessionId.value,
  (id) => emit("session-change", id),
);

function handleSend() {
  const text = input.value.trim();
  if (!text || chat.isStreaming.value) return;
  input.value = "";
  emit("send", text);
  chat.sendMessage(text);
  if (inputEl.value) inputEl.value.style.height = "auto";
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    handleSend();
  }
}

function handleSuggestion(text: string) {
  input.value = "";
  emit("send", text);
  chat.sendMessage(text);
}

function handleInput(text: string) {
  chat.sendMessage(text);
}

function handleTextareaInput(e: Event) {
  const el = e.target as HTMLTextAreaElement;
  input.value = el.value;
  el.style.height = "auto";
  el.style.height = `${Math.min(el.scrollHeight, 160)}px`;
}
</script>

<template>
  <div
    class="arp-chat"
    :style="{
      display: 'flex',
      height: '100%',
      background: isDark ? '#09090b' : '#ffffff',
      color: isDark ? '#e0e0e0' : '#1a1a1a',
      fontFamily: `-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif`,
    }"
  >
    <!-- Main chat area -->
    <div style="flex: 1; display: flex; flex-direction: column; min-width: 0">
      <!-- Header -->
      <div
        class="arp-chat-header"
        :style="{
          padding: '12px 20px',
          borderBottom: `1px solid ${isDark ? '#1a1a2e' : '#e5e5e5'}`,
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
        }"
      >
        <img v-if="logo" :src="logo" alt="" style="height: 24px; width: 24px" />
        <span style="font-weight: 600; font-size: 15px">{{ title }}</span>
        <span
          v-if="chat.isConnected.value"
          style="
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: #22c55e;
            margin-left: 4px;
          "
        />
      </div>

      <!-- Messages area -->
      <div
        class="arp-chat-messages"
        style="flex: 1; overflow: auto; padding: 0 20px"
      >
        <!-- Welcome screen -->
        <div
          v-if="showWelcome"
          style="
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100%;
            gap: 16px;
            text-align: center;
          "
        >
          <div :style="{ fontSize: '28px', fontWeight: 700, color: accentColor }">
            {{ title }}
          </div>
          <div
            v-if="description"
            :style="{ color: isDark ? '#888' : '#666', maxWidth: '480px' }"
          >
            {{ description }}
          </div>
          <div
            v-if="suggestions && suggestions.length > 0"
            style="
              display: flex;
              flex-wrap: wrap;
              gap: 8px;
              justify-content: center;
              margin-top: 8px;
            "
          >
            <button
              v-for="(s, i) in suggestions"
              :key="i"
              :style="{
                background: isDark ? '#1a1a2e' : '#f0f0f0',
                color: isDark ? '#ccc' : '#333',
                border: `1px solid ${isDark ? '#2a2a3e' : '#ddd'}`,
                borderRadius: '20px',
                padding: '8px 16px',
                cursor: 'pointer',
                fontSize: '13px',
                transition: 'background 0.15s',
              }"
              @click="handleSuggestion(s)"
            >
              {{ s }}
            </button>
          </div>
        </div>

        <!-- Message list -->
        <ArpMessage
          v-for="msg in chat.messages.value"
          :key="msg.id"
          :role="msg.role"
          :content="msg.content"
          :avatar="avatar"
          :file="msg.file"
          :ui-events="msg.uiEvents"
          :restored="msg.restored"
          @input="handleInput"
        />

        <!-- Typing indicator -->
        <div
          v-if="
            chat.isStreaming.value &&
            chat.messages.value.length > 0 &&
            !chat.messages.value[chat.messages.value.length - 1]?.content
          "
          style="padding: 8px 0; color: #666; font-size: 13px"
        >
          Thinking...
        </div>

        <div ref="messagesEnd" />
      </div>

      <!-- Input area -->
      <div
        class="arp-chat-input"
        :style="{
          padding: '12px 20px',
          borderTop: `1px solid ${isDark ? '#1a1a2e' : '#e5e5e5'}`,
        }"
      >
        <div
          :style="{
            display: 'flex',
            gap: '10px',
            alignItems: 'flex-end',
            background: isDark ? '#111118' : '#f8f8f8',
            borderRadius: '12px',
            padding: '8px 12px',
            border: `1px solid ${isDark ? '#2a2a3e' : '#ddd'}`,
          }"
        >
          <textarea
            ref="inputEl"
            :value="input"
            placeholder="Type a message..."
            :disabled="chat.isStreaming.value"
            rows="1"
            :style="{
              flex: 1,
              background: 'transparent',
              border: 'none',
              outline: 'none',
              color: isDark ? '#e0e0e0' : '#1a1a1a',
              fontSize: '14px',
              lineHeight: 1.5,
              resize: 'none',
              fontFamily: 'inherit',
              maxHeight: '160px',
            }"
            @input="handleTextareaInput"
            @keydown="handleKeyDown"
          />
          <button
            :disabled="chat.isStreaming.value || !input.trim()"
            :style="{
              background: accentColor,
              color: '#000',
              border: 'none',
              borderRadius: '8px',
              padding: '6px 14px',
              cursor: chat.isStreaming.value || !input.trim() ? 'default' : 'pointer',
              opacity: chat.isStreaming.value || !input.trim() ? 0.4 : 1,
              fontWeight: 600,
              fontSize: '13px',
              flexShrink: 0,
              transition: 'opacity 0.15s',
            }"
            @click="handleSend"
          >
            Send
          </button>
        </div>
      </div>
    </div>

    <!-- Activity panel -->
    <ArpActivityPanel
      v-if="showActivityPanel"
      :tool-cards="chat.toolCards.value"
      :open="panelOpen"
      @toggle="panelOpen = !panelOpen"
    />
  </div>
</template>
