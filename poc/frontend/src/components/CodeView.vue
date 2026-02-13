<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchHairaSource } from '../services/chat'

const source = ref('')
const loading = ref(true)

onMounted(async () => {
  try {
    source.value = await fetchHairaSource()
  } catch {
    source.value = '// Could not load source'
  }
  loading.value = false
})

function highlight(code: string): string {
  // Simple syntax highlighting for Haira
  return code
    // Triple-quote strings (must be first)
    .replace(/("""[\s\S]*?""")/g, '<span class="hl-string">$1</span>')
    // Single-line comments
    .replace(/(\/\/.*)/g, '<span class="hl-comment">$1</span>')
    // Strings
    .replace(/("(?:[^"\\]|\\.)*")/g, '<span class="hl-string">$1</span>')
    // Agentic keywords
    .replace(/\b(provider|tool|agent|workflow)\b/g, '<span class="hl-agent">$1</span>')
    // Core keywords
    .replace(/\b(fn|import|if|else|for|return|match|struct|enum|spawn|nil)\b/g, '<span class="hl-keyword">$1</span>')
    // Types
    .replace(/\b(string|int|float|bool|stream)\b/g, '<span class="hl-type">$1</span>')
    // Decorators
    .replace(/(@\w+)/g, '<span class="hl-decorator">$1</span>')
    // env()
    .replace(/\b(env)\(/g, '<span class="hl-fn">$1</span>(')
}
</script>

<template>
  <div class="code-view">
    <div class="code-header">
      <span class="file-icon">&#128196;</span>
      <span class="file-name">main.haira</span>
      <span class="file-badge">Live</span>
    </div>
    <div class="code-content">
      <div v-if="loading" class="code-loading">Loading source...</div>
      <pre v-else><code v-html="highlight(source)"></code></pre>
    </div>
    <div class="code-footer">
      <span>Haira Runtime v0.1 &mdash; Agentic Orchestration Language</span>
    </div>
  </div>
</template>

<style scoped>
.code-view { display: flex; flex-direction: column; height: 100%; }

.code-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border);
  font-size: 13px;
  flex-shrink: 0;
}
.file-icon { font-size: 16px; }
.file-name { color: var(--text); font-weight: 600; }
.file-badge {
  margin-left: auto;
  background: #1a7f37;
  color: white;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
}

.code-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.code-loading { color: var(--text-muted); font-size: 14px; }

pre {
  margin: 0;
  font-family: 'SF Mono', 'Fira Code', 'Cascadia Code', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre;
  tab-size: 4;
}

code { color: var(--text); }

.code-footer {
  padding: 8px 16px;
  border-top: 1px solid var(--border);
  font-size: 12px;
  color: var(--text-muted);
  flex-shrink: 0;
}

:deep(.hl-agent) { color: #d2a8ff; font-weight: 700; }
:deep(.hl-keyword) { color: #ff7b72; }
:deep(.hl-type) { color: #79c0ff; }
:deep(.hl-string) { color: #a5d6ff; }
:deep(.hl-comment) { color: #8b949e; font-style: italic; }
:deep(.hl-decorator) { color: #ffa657; font-weight: 600; }
:deep(.hl-fn) { color: #d2a8ff; }
</style>
