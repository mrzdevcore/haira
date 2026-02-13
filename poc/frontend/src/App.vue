<script setup lang="ts">
import ChatView from './components/ChatView.vue'
import CodeView from './components/CodeView.vue'
</script>

<template>
  <div class="app">
    <header class="header">
      <div class="header-left">
        <h1 class="logo">Craftify</h1>
        <span class="tagline">Powered by <strong>Haira</strong></span>
      </div>
      <div class="header-right">
        <span class="badge">POC</span>
      </div>
    </header>
    <main class="main">
      <div class="panel chat-panel">
        <ChatView />
      </div>
      <div class="panel code-panel">
        <CodeView />
      </div>
    </main>
  </div>
</template>

<style>
:root {
  --bg: #0d1117;
  --bg-secondary: #161b22;
  --bg-tertiary: #21262d;
  --border: #30363d;
  --text: #e6edf3;
  --text-muted: #8b949e;
  --accent: #58a6ff;
  --accent-green: #3fb950;
  --user-bubble: #1f6feb;
  --assistant-bubble: #21262d;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  background: var(--bg);
  color: var(--text);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  height: 100vh;
  overflow: hidden;
}

#app { height: 100vh; display: flex; flex-direction: column; }

.app { display: flex; flex-direction: column; height: 100vh; }

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 24px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.header-left { display: flex; align-items: center; gap: 16px; }

.logo {
  font-size: 24px;
  font-weight: 700;
  background: linear-gradient(135deg, #58a6ff, #3fb950);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.tagline { color: var(--text-muted); font-size: 14px; }
.tagline strong { color: var(--accent); }

.badge {
  background: var(--accent);
  color: var(--bg);
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.main {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.panel { overflow: hidden; display: flex; flex-direction: column; }

.chat-panel {
  flex: 6;
  border-right: 1px solid var(--border);
}

.code-panel {
  flex: 4;
  background: var(--bg-secondary);
}
</style>
