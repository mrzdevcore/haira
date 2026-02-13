<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { createConversation, sendMessage, type ChatMessage } from '../services/chat'

const messages = ref<ChatMessage[]>([])
const input = ref('')
const loading = ref(false)
const conversationId = ref<string | null>(null)
const userId = ref('demo-user-' + Math.random().toString(36).slice(2, 8))
const messagesContainer = ref<HTMLElement | null>(null)

async function scrollToBottom() {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

async function send() {
  const text = input.value.trim()
  if (!text || loading.value) return

  const userMsg: ChatMessage = {
    id: Date.now().toString(),
    role: 'user',
    content: text,
    timestamp: new Date(),
  }
  messages.value.push(userMsg)
  input.value = ''
  loading.value = true
  await scrollToBottom()

  try {
    if (!conversationId.value) {
      const result = await createConversation(userId.value, text)
      conversationId.value = result.conversationId
      messages.value.push(result.reply)
    } else {
      const reply = await sendMessage(userId.value, conversationId.value, text)
      messages.value.push(reply)
    }
  } catch (err: any) {
    messages.value.push({
      id: Date.now().toString(),
      role: 'assistant',
      content: 'Sorry, something went wrong. Please try again.',
      timestamp: new Date(),
    })
  }

  loading.value = false
  await scrollToBottom()
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}

function newChat() {
  messages.value = []
  conversationId.value = null
  input.value = ''
}

function formatContent(content: string): string {
  // Basic markdown-like formatting
  return content
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\n/g, '<br>')
}
</script>

<template>
  <div class="chat">
    <div class="chat-header">
      <span>Chat with BrewingAdvisor</span>
      <button class="new-chat-btn" @click="newChat">+ New</button>
    </div>

    <div class="messages" ref="messagesContainer">
      <div v-if="messages.length === 0" class="welcome">
        <div class="welcome-icon">&#127866;</div>
        <h2>Welcome to Craftify</h2>
        <p>I'm your AI brewing advisor. Ask me about beer recipes, ingredients, or costs.</p>
        <div class="suggestions">
          <button @click="input = 'Find me a good IPA recipe'; send()">Find me a good IPA recipe</button>
          <button @click="input = 'What malts work best for a porter?'; send()">What malts for a porter?</button>
          <button @click="input = 'How much to brew 100L of pale ale?'; send()">Cost for 100L pale ale?</button>
        </div>
      </div>

      <div
        v-for="msg in messages"
        :key="msg.id"
        :class="['message', msg.role]"
      >
        <div class="avatar">{{ msg.role === 'user' ? 'You' : 'AI' }}</div>
        <div class="bubble" v-html="formatContent(msg.content)"></div>
      </div>

      <div v-if="loading" class="message assistant">
        <div class="avatar">AI</div>
        <div class="bubble loading-bubble">
          <span class="dot"></span>
          <span class="dot"></span>
          <span class="dot"></span>
        </div>
      </div>
    </div>

    <div class="input-area">
      <textarea
        v-model="input"
        @keydown="handleKeydown"
        placeholder="Ask about recipes, malts, or brewing costs..."
        :disabled="loading"
        rows="1"
      ></textarea>
      <button @click="send" :disabled="!input.trim() || loading" class="send-btn">
        Send
      </button>
    </div>
  </div>
</template>

<style scoped>
.chat { display: flex; flex-direction: column; height: 100%; }

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border);
  font-weight: 600;
  font-size: 14px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.new-chat-btn {
  background: var(--bg-tertiary);
  color: var(--text);
  border: 1px solid var(--border);
  padding: 4px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
}
.new-chat-btn:hover { background: var(--border); }

.messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.welcome {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  gap: 12px;
  color: var(--text-muted);
}
.welcome-icon { font-size: 48px; }
.welcome h2 { color: var(--text); font-size: 20px; }
.welcome p { font-size: 14px; max-width: 400px; text-align: center; }

.suggestions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 8px; justify-content: center; }
.suggestions button {
  background: var(--bg-tertiary);
  color: var(--text);
  border: 1px solid var(--border);
  padding: 8px 16px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.15s;
}
.suggestions button:hover { border-color: var(--accent); color: var(--accent); }

.message { display: flex; gap: 12px; max-width: 85%; }
.message.user { align-self: flex-end; flex-direction: row-reverse; }
.message.assistant { align-self: flex-start; }

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}
.message.user .avatar { background: var(--user-bubble); color: white; }
.message.assistant .avatar { background: var(--accent-green); color: var(--bg); }

.bubble {
  padding: 10px 16px;
  border-radius: 16px;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
}
.message.user .bubble {
  background: var(--user-bubble);
  color: white;
  border-bottom-right-radius: 4px;
}
.message.assistant .bubble {
  background: var(--assistant-bubble);
  border: 1px solid var(--border);
  border-bottom-left-radius: 4px;
}

.loading-bubble { display: flex; gap: 4px; padding: 14px 20px; }
.dot {
  width: 8px;
  height: 8px;
  background: var(--text-muted);
  border-radius: 50%;
  animation: bounce 1.4s infinite;
}
.dot:nth-child(2) { animation-delay: 0.2s; }
.dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce {
  0%, 80%, 100% { transform: translateY(0); }
  40% { transform: translateY(-8px); }
}

.input-area {
  display: flex;
  gap: 8px;
  padding: 16px 20px;
  border-top: 1px solid var(--border);
  background: var(--bg-secondary);
  flex-shrink: 0;
}

textarea {
  flex: 1;
  background: var(--bg-tertiary);
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 10px 16px;
  font-size: 14px;
  font-family: inherit;
  resize: none;
  outline: none;
}
textarea:focus { border-color: var(--accent); }
textarea::placeholder { color: var(--text-muted); }

.send-btn {
  background: var(--accent);
  color: var(--bg);
  border: none;
  padding: 10px 20px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  transition: opacity 0.15s;
}
.send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.send-btn:not(:disabled):hover { opacity: 0.85; }
</style>
