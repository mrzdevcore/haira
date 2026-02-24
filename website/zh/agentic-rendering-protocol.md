---
layout: page
title: ARP — 代理渲染协议
description: 一种与传输方式无关的协议，用于 AI 代理与渲染界面之间的通信。
---

<div class="blog-page">
<div class="blog-hero">
  <div class="blog-hero-glow"></div>
  <div class="blog-hero-inner">
    <div class="blog-badge">协议规范</div>
    <h1 class="blog-title">代理渲染协议</h1>
    <p class="blog-subtitle">一种与传输方式无关的双向协议，用于 AI 代理与渲染界面之间的通信。一个代理，适配所有界面。</p>
    <div class="blog-meta">
      <div class="blog-meta-item">v0.1 草案</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">开放标准</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">CC BY 4.0</div>
    </div>
  </div>
</div>

<div class="blog-body">

<div class="blog-section">

## 问题所在

AI 代理会生成结构化输出——表格、表单、图表、状态消息、代码块——但这些输出的渲染方式与传输层和前端框架紧密耦合。为 Web 聊天构建的代理无法在桌面应用中渲染。通过 SSE 流传输的代理无法服务于期望不同协议的移动客户端。

每接入一个新的渲染界面，就需要编写一套定制集成代码。

</div>

<div class="blog-section">

## 解决方案

ARP 定义的是一套**协议**——而非框架，也非库。就像 Wayland 将应用程序从显示服务器中解耦一样，ARP 将代理从渲染器中解耦。

代理决定**显示什么**。渲染器决定**如何显示**。

<div class="blog-card">
<div class="blog-card-header">架构</div>

```
Agent (backend)                              Renderer (web, CLI, mobile)
───────────────                              ──────────────────────────
Owns surfaces         ── render / delta ──►  Owns display
Owns UI state         ◄── input events ────  Owns input routing
```

</div>
</div>

<div class="blog-section">

## 设计原则

<div class="principle-grid">
  <div class="principle-item">
    <div class="principle-num">01</div>
    <div class="principle-text">
      <div class="principle-title">无中间层</div>
      <div class="principle-desc">代理直接发送渲染指令。渲染器直接发送输入。无需中间框架。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">02</div>
    <div class="principle-text">
      <div class="principle-title">异步 &amp; 非阻塞</div>
      <div class="principle-desc">所有消息均为异步。任何一方都不会阻塞等待响应。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">03</div>
    <div class="principle-text">
      <div class="principle-title">每帧皆完美</div>
      <div class="principle-desc">原子提交。渲染状态在待定缓冲区中累积，原子性地应用。无闪烁。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">04</div>
    <div class="principle-text">
      <div class="principle-title">能力驱动</div>
      <div class="principle-desc">渲染器声明其支持的能力，代理随之适配。CLI 获得表格，Web 获得一切。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">05</div>
    <div class="principle-text">
      <div class="principle-title">传输无关</div>
      <div class="principle-desc">逻辑消息，而非线格式。WebSocket、SSE、gRPC、Unix 套接字、stdio。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">06</div>
    <div class="principle-text">
      <div class="principle-title">类型化组件</div>
      <div class="principle-desc">代理发出类型化描述符，例如带有 <code>headers</code> 和 <code>rows</code> 的 <code>table</code>——而非 HTML。</div>
    </div>
  </div>
</div>

</div>

<div class="blog-section">

## 协议消息

每条 ARP 消息都是 JSON 格式，至少包含 `{ v: 1, type: "<type>" }`。

<div class="blog-card">
<div class="blog-card-header">服务端到客户端</div>

| 类型 | 用途 |
|------|---------|
| `hello` | 连接时的能力握手 |
| `delta` | 增量文本块 |
| `tool_start` | 工具开始执行 |
| `tool_end` | 工具执行完毕 |
| `render` | 生成式 UI 组件 |
| `patch` | 组件增量更新 |
| `error` | 错误事件 |
| `commit` | 流传输完成 |

</div>

<div class="blog-card">
<div class="blog-card-header">客户端到服务端</div>

| 类型 | 输入类型 | 用途 |
|------|-----------|---------|
| `input` | `text` | 用户文本消息 |
| `input` | `action` | 按钮点击 / UI 操作 |
| `input` | `form_submit` | 表单提交 |

</div>
</div>

<div class="blog-section">

## 14 个内置组件

每个符合 ARP 规范的渲染器至少须支持 `text`。组件声明回退链——`chart` 回退到 `table`，`table` 回退到 `text`。

<div class="component-pills">
  <div class="pill">text</div>
  <div class="pill">markdown</div>
  <div class="pill">status-card</div>
  <div class="pill">table</div>
  <div class="pill">code-block</div>
  <div class="pill">diff</div>
  <div class="pill">key-value</div>
  <div class="pill">progress</div>
  <div class="pill">chart</div>
  <div class="pill">form</div>
  <div class="pill">confirm</div>
  <div class="pill">choices</div>
  <div class="pill">product-cards</div>
  <div class="pill">image</div>
</div>

</div>

<div class="blog-section">

## 传输方式

<div class="transport-grid">
  <div class="transport-card transport-active">
    <div class="transport-label">可用</div>
    <div class="transport-name">WebSocket</div>
    <div class="transport-path"><code>/_arp/v1</code></div>
    <div class="transport-desc">主要传输方式。持久双向连接，支持自动重连。</div>
  </div>
  <div class="transport-card transport-active">
    <div class="transport-label">可用</div>
    <div class="transport-name">SSE</div>
    <div class="transport-path">Server-Sent Events</div>
    <div class="transport-desc">备用传输方式。文件上传需通过 multipart/form-data 时使用。</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">计划中</div>
    <div class="transport-name">gRPC</div>
    <div class="transport-path">高性能</div>
    <div class="transport-desc">适用于原生桌面和移动应用程序。</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">计划中</div>
    <div class="transport-name">stdio</div>
    <div class="transport-path">NDJSON 分帧</div>
    <div class="transport-desc">适用于 CLI 渲染器和基于管道的集成。</div>
  </div>
</div>

</div>

<div class="blog-section">

## 客户端 SDK

<div class="blog-card">
<div class="blog-card-header">@haira/arp &mdash; 核心库（零依赖）</div>

```typescript
import { ArpClient } from '@haira/arp'

const client = new ArpClient('ws://localhost:8080/_arp/v1', {
  onDelta: (text) => appendToChat(text),
  onRender: (event) => renderComponent(event.component, event.props),
  onDone: () => markStreamComplete(),
})

client.connect()
client.sendText('Show me the sales data')
```

</div>

<div class="blog-card">
<div class="blog-card-header">@haira/arp-react &mdash; 即插即用聊天 UI</div>

```tsx
import { ArpChat } from '@haira/arp-react'

function App() {
  return (
    <ArpChat
      url="ws://localhost:8080/_arp/v1"
      theme="dark"
      title="Data Explorer"
    />
  )
}
```

</div>

同样可用：**`@haira/arp-vue`**（适用于 Vue 3）和 **`github.com/haira-lang/arp-go`**（适用于 Go 后端）。

</div>

<div class="blog-section">

## Haira 集成

每个 Haira 服务器原生支持 ARP，无需任何配置。

```haira
import "ui"

tool query_database(query: string) -> any {
    """Executes a SQL query and displays results."""
    rows, err = postgres.query(db, query)
    if err != nil {
        return ui.status_card("error", "Query Failed", conv.to_string(err))
    }
    return ui.table("Query Results", headers, rows)
}

agent DataExplorer {
    provider: OpenAI
    tools: [query_database]
    ui: ui
}

@webhook("/chat")
workflow Chat(message: string, session_id: string) -> stream {
    return DataExplorer.stream(message, session: session_id)
}
```

</div>

<div class="blog-section">

## 扩展生命周期

新组件遵循受 Wayland 启发的三阶段生命周期：

<div class="lifecycle-steps">
  <div class="lifecycle-step">
    <div class="lifecycle-phase">1. 实验阶段</div>
    <div class="lifecycle-prefix"><code>x-vendor-name</code></div>
    <div class="lifecycle-desc">由厂商创建。允许破坏性变更。</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">2. 预发布阶段</div>
    <div class="lifecycle-prefix"><code>s-name</code></div>
    <div class="lifecycle-desc">需要至少 2 个渲染器实现。经过治理审查。</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">3. 核心阶段</div>
    <div class="lifecycle-prefix">无前缀</div>
    <div class="lifecycle-desc">纳入 ARP 规范。仅允许向后兼容的变更。</div>
  </div>
</div>

</div>

<div class="blog-section">

## 合规级别

| 级别 | 必需组件 | 适用场景 |
|-------|-------------------|--------|
| 最小 | `text` + 文本输入 | 语音助手、IoT |
| 基础 | text、table、form、confirm、choices | CLI 终端 |
| 标准 | 所有核心组件，完整对象模型 | Web / 桌面 |
| 完整 | 标准级 + 流传输、多界面、文件上传 | 富 Web 应用 |

</div>

<div class="blog-cta">
  <a class="blog-cta-btn" href="/zh/docs/agentic/arp">阅读完整参考文档</a>
</div>

</div>
</div>

<style>
.blog-page {
  --gold: #E8A317;
  --gold-light: #F0BD4F;
  --gold-glow: #FDE68A;
}

/* ── Hero ── */
.blog-hero {
  position: relative;
  text-align: center;
  padding: 6rem 2rem 5rem;
  overflow: hidden;
}
.blog-hero-glow {
  position: absolute;
  top: -150px;
  left: 50%;
  transform: translateX(-50%);
  width: 800px;
  height: 500px;
  background: radial-gradient(ellipse, rgba(232, 163, 23, 0.1) 0%, transparent 70%);
  pointer-events: none;
  z-index: 0;
}
.blog-hero-inner {
  position: relative;
  z-index: 1;
  max-width: 960px;
  margin: 0 auto;
}
.blog-badge {
  display: inline-block;
  padding: 0.375rem 1.125rem;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  background: rgba(232, 163, 23, 0.1);
  color: var(--gold);
  border: 1px solid rgba(232, 163, 23, 0.18);
  margin-bottom: 2rem;
}
.blog-title {
  font-size: clamp(2.5rem, 5.5vw, 4rem);
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1.2;
  background: linear-gradient(135deg, var(--gold) 0%, var(--gold-light) 50%, var(--gold-glow) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 1.5rem;
  padding: 0 1rem 0.1em;
}
.blog-subtitle {
  font-size: 1.25rem;
  color: var(--vp-c-text-2);
  max-width: 600px;
  margin: 0 auto 2rem;
  line-height: 1.7;
}
.blog-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.875rem;
  font-size: 0.875rem;
  color: var(--vp-c-text-3);
}
.blog-meta-sep {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--vp-c-text-3);
  opacity: 0.4;
}

/* ── Body ── */
.blog-body {
  max-width: 960px;
  margin: 0 auto;
  padding: 0 2rem 6rem;
}
.blog-section {
  padding: 3rem 0;
  border-bottom: 1px solid var(--vp-c-divider);
}
.blog-section:last-child {
  border-bottom: none;
}
.blog-body h2 {
  font-size: 1.875rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  margin: 0 0 1.25rem;
  color: var(--vp-c-text-1);
}
.blog-body h3 {
  font-size: 1.125rem;
  font-weight: 600;
  margin: 1.75rem 0 0.75rem;
  color: var(--vp-c-text-1);
}
.blog-body p {
  margin: 1rem 0;
  color: var(--vp-c-text-2);
  line-height: 1.8;
  font-size: 1.0625rem;
}
.blog-body a {
  color: var(--gold);
  text-decoration: none;
  font-weight: 500;
}
.blog-body a:hover {
  color: var(--gold-light);
}
.blog-body table {
  width: 100%;
  border-collapse: collapse;
  margin: 1.25rem 0;
  font-size: 0.9375rem;
}
.blog-body th {
  text-align: left;
  padding: 0.75rem 1.25rem;
  border-bottom: 2px solid var(--vp-c-divider);
  font-weight: 600;
  color: var(--vp-c-text-1);
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.blog-body td {
  padding: 0.625rem 1.25rem;
  border-bottom: 1px solid var(--vp-c-divider);
  color: var(--vp-c-text-2);
}

/* ── Cards ── */
.blog-card {
  border: 1px solid var(--vp-c-divider);
  border-radius: 14px;
  overflow: hidden;
  margin: 1.5rem 0;
  background: var(--vp-c-bg-soft);
}
.blog-card-header {
  padding: 0.875rem 1.5rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--vp-c-text-2);
  border-bottom: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg-alt);
  letter-spacing: 0.01em;
}
.blog-card table { margin: 0; }
.blog-card th:first-child,
.blog-card td:first-child { padding-left: 1.5rem; }
.blog-card div[class*="language-"] {
  margin: 0 !important;
  border-radius: 0 !important;
}

/* ── Principle Grid ── */
.principle-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1px;
  background: var(--vp-c-divider);
  border: 1px solid var(--vp-c-divider);
  border-radius: 14px;
  overflow: hidden;
  margin: 1.5rem 0;
}
.principle-item {
  display: flex;
  gap: 1.125rem;
  padding: 1.5rem;
  background: var(--vp-c-bg-soft);
}
.principle-num {
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--gold);
  font-family: 'JetBrains Mono', monospace;
  line-height: 1.6;
  opacity: 0.6;
}
.principle-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
  margin-bottom: 0.375rem;
}
.principle-desc {
  font-size: 0.9375rem;
  color: var(--vp-c-text-3);
  line-height: 1.6;
}

/* ── Component Pills ── */
.component-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.625rem;
  margin: 1.5rem 0;
}
.component-pills .pill {
  padding: 0.5rem 1.125rem;
  border-radius: 999px;
  font-size: 0.9375rem;
  font-weight: 500;
  font-family: 'JetBrains Mono', monospace;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  color: var(--vp-c-text-2);
  transition: border-color 0.15s, color 0.15s;
}
.component-pills .pill:hover {
  border-color: var(--gold);
  color: var(--gold);
}

/* ── Transport Grid ── */
.transport-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin: 1.5rem 0;
}
.transport-card {
  padding: 1.5rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background: var(--vp-c-bg-soft);
}
.transport-card.transport-active {
  border-color: rgba(232, 163, 23, 0.3);
}
.transport-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--vp-c-text-3);
  margin-bottom: 0.5rem;
}
.transport-active .transport-label { color: var(--gold); }
.transport-name {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin-bottom: 0.25rem;
}
.transport-path {
  font-size: 0.8125rem;
  color: var(--vp-c-text-3);
  margin-bottom: 0.625rem;
}
.transport-desc {
  font-size: 0.9375rem;
  color: var(--vp-c-text-3);
  line-height: 1.6;
}

/* ── Lifecycle ── */
.lifecycle-steps {
  display: flex;
  gap: 1rem;
  margin: 1.5rem 0;
}
.lifecycle-step {
  flex: 1;
  padding: 1.5rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background: var(--vp-c-bg-soft);
}
.lifecycle-phase {
  font-size: 1rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin-bottom: 0.375rem;
}
.lifecycle-prefix {
  font-size: 0.8125rem;
  color: var(--gold);
  margin-bottom: 0.625rem;
}
.lifecycle-desc {
  font-size: 0.9375rem;
  color: var(--vp-c-text-3);
  line-height: 1.6;
}

/* ── CTA ── */
.blog-cta {
  text-align: center;
  padding: 4rem 0 2rem;
}
.blog-cta-btn {
  display: inline-block;
  padding: 0.875rem 2.5rem;
  border-radius: 12px;
  font-weight: 700;
  font-size: 1rem;
  color: #1a1a2e !important;
  background: linear-gradient(135deg, var(--gold) 0%, var(--gold-light) 100%);
  text-decoration: none !important;
  transition: opacity 0.15s, transform 0.15s;
}
.blog-cta-btn:hover {
  opacity: 0.9;
  transform: translateY(-1px);
  color: #1a1a2e !important;
}

/* ── Overflow ── */
.blog-page {
  overflow-x: hidden;
}
.blog-body div[class*="language-"] {
  overflow-x: auto;
}
.blog-body table {
  display: block;
  overflow-x: auto;
}
.blog-card {
  overflow-x: auto;
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .blog-hero { padding: 4rem 1.5rem 3rem; }
  .blog-title { font-size: 2rem; }
  .blog-subtitle { font-size: 1.0625rem; }
  .blog-body { padding: 0 1.25rem 4rem; }
  .blog-body p { font-size: 1rem; }
  .principle-grid { grid-template-columns: 1fr; }
  .transport-grid { grid-template-columns: 1fr; }
  .lifecycle-steps { flex-direction: column; }
}
</style>
