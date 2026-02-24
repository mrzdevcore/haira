---
layout: page
title: ARP — 에이전트 렌더링 프로토콜
description: AI 에이전트와 렌더링 표면 간의 통신을 위한 전송 계층 독립적 프로토콜.
---

<div class="blog-page">
<div class="blog-hero">
  <div class="blog-hero-glow"></div>
  <div class="blog-hero-inner">
    <div class="blog-badge">프로토콜 사양</div>
    <h1 class="blog-title">에이전트 렌더링 프로토콜</h1>
    <p class="blog-subtitle">AI 에이전트와 렌더링 표면 간의 양방향 통신을 위한 전송 계층 독립적 프로토콜. 하나의 에이전트, 모든 화면.</p>
    <div class="blog-meta">
      <div class="blog-meta-item">v0.1 초안</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">오픈 표준</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">CC BY 4.0</div>
    </div>
  </div>
</div>

<div class="blog-body">

<div class="blog-section">

## 문제

AI 에이전트는 구조화된 출력 — 표, 폼, 차트, 상태 메시지, 코드 블록 — 을 생성하지만, 해당 출력의 렌더링은 전송 계층 및 프론트엔드 프레임워크에 긴밀하게 결합되어 있습니다. 웹 채팅용으로 구축된 에이전트는 데스크톱 앱에서 렌더링할 수 없습니다. SSE 스트림 뒤에 있는 에이전트는 다른 프로토콜을 기대하는 모바일 클라이언트에 서비스할 수 없습니다.

새로운 렌더링 표면마다 맞춤형 통합 코드가 필요합니다.

</div>

<div class="blog-section">

## 해결책

ARP는 **프로토콜**을 정의합니다 — 프레임워크도, 라이브러리도 아닙니다. Wayland가 애플리케이션을 디스플레이 서버로부터 분리한 것처럼, ARP는 에이전트를 렌더러로부터 분리합니다.

에이전트는 **무엇을** 표시할지 결정합니다. 렌더러는 **어떻게** 표시할지 결정합니다.

<div class="blog-card">
<div class="blog-card-header">아키텍처</div>

```
Agent (backend)                              Renderer (web, CLI, mobile)
───────────────                              ──────────────────────────
Owns surfaces         ── render / delta ──►  Owns display
Owns UI state         ◄── input events ────  Owns input routing
```

</div>
</div>

<div class="blog-section">

## 설계 원칙

<div class="principle-grid">
  <div class="principle-item">
    <div class="principle-num">01</div>
    <div class="principle-text">
      <div class="principle-title">중간자 없음</div>
      <div class="principle-desc">에이전트가 렌더 명령을 직접 전송합니다. 렌더러가 입력을 직접 전송합니다. 중간 프레임워크 없음.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">02</div>
    <div class="principle-text">
      <div class="principle-title">비동기 &amp; 논블로킹</div>
      <div class="principle-desc">모든 메시지는 비동기입니다. 어느 쪽도 응답을 기다리며 블로킹하지 않습니다.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">03</div>
    <div class="principle-text">
      <div class="principle-title">모든 프레임은 완벽</div>
      <div class="principle-desc">원자적 커밋. 렌더 상태는 대기 버퍼에 누적되고 원자적으로 적용됩니다. 깜빡임 없음.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">04</div>
    <div class="principle-text">
      <div class="principle-title">기능 기반</div>
      <div class="principle-desc">렌더러가 지원하는 기능을 선언합니다. 에이전트가 적응합니다. CLI는 표를, 웹은 모든 것을 받습니다.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">05</div>
    <div class="principle-text">
      <div class="principle-title">전송 계층 독립</div>
      <div class="principle-desc">논리적 메시지, 와이어 포맷이 아닙니다. WebSocket, SSE, gRPC, Unix 소켓, stdio.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">06</div>
    <div class="principle-text">
      <div class="principle-title">타입이 있는 컴포넌트</div>
      <div class="principle-desc">에이전트는 HTML이 아닌 <code>headers</code>와 <code>rows</code>가 있는 <code>table</code> 같은 타입 디스크립터를 내보냅니다.</div>
    </div>
  </div>
</div>

</div>

<div class="blog-section">

## 프로토콜 메시지

모든 ARP 메시지는 최소한 `{ v: 1, type: "<type>" }`을 포함하는 JSON입니다.

<div class="blog-card">
<div class="blog-card-header">서버 → 클라이언트</div>

| 타입 | 목적 |
|------|---------|
| `hello` | 연결 시 기능 핸드셰이크 |
| `delta` | 증분 텍스트 청크 |
| `tool_start` | 도구 실행 시작 |
| `tool_end` | 도구 실행 완료 |
| `render` | 생성형 UI 컴포넌트 |
| `patch` | 증분 컴포넌트 업데이트 |
| `error` | 오류 이벤트 |
| `commit` | 스트림 완료 |

</div>

<div class="blog-card">
<div class="blog-card-header">클라이언트 → 서버</div>

| 타입 | 입력 타입 | 목적 |
|------|-----------|---------|
| `input` | `text` | 사용자 텍스트 메시지 |
| `input` | `action` | 버튼 클릭 / UI 액션 |
| `input` | `form_submit` | 폼 제출 |

</div>
</div>

<div class="blog-section">

## 14가지 기본 컴포넌트

ARP 적합 렌더러는 최소한 `text`를 지원해야 합니다. 컴포넌트는 폴백 체인을 선언합니다 — `chart`는 `table`로, `table`은 `text`로 폴백됩니다.

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

## 전송 계층

<div class="transport-grid">
  <div class="transport-card transport-active">
    <div class="transport-label">사용 가능</div>
    <div class="transport-name">WebSocket</div>
    <div class="transport-path"><code>/_arp/v1</code></div>
    <div class="transport-desc">기본 전송 계층. 자동 재연결이 있는 지속적 양방향 연결.</div>
  </div>
  <div class="transport-card transport-active">
    <div class="transport-label">사용 가능</div>
    <div class="transport-name">SSE</div>
    <div class="transport-path">Server-Sent Events</div>
    <div class="transport-desc">폴백 전송 계층. multipart/form-data를 통한 파일 업로드에 필수.</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">예정</div>
    <div class="transport-name">gRPC</div>
    <div class="transport-path">고성능</div>
    <div class="transport-desc">네이티브 데스크톱 및 모바일 애플리케이션용.</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">예정</div>
    <div class="transport-name">stdio</div>
    <div class="transport-path">NDJSON 프레이밍</div>
    <div class="transport-desc">CLI 렌더러 및 파이프 기반 통합용.</div>
  </div>
</div>

</div>

<div class="blog-section">

## 클라이언트 SDK

<div class="blog-card">
<div class="blog-card-header">@haira/arp &mdash; 코어 (의존성 없음)</div>

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
<div class="blog-card-header">@haira/arp-react &mdash; 즉시 사용 가능한 채팅 UI</div>

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

Vue 3용 **`@haira/arp-vue`** 및 Go 백엔드용 **`github.com/haira-lang/arp-go`**도 사용 가능합니다.

</div>

<div class="blog-section">

## Haira 통합

모든 Haira 서버는 기본적으로 ARP를 지원합니다. 별도의 설정이 필요 없습니다.

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

## 확장 수명 주기

새 컴포넌트는 Wayland에서 영감을 받은 3단계 수명 주기를 따릅니다:

<div class="lifecycle-steps">
  <div class="lifecycle-step">
    <div class="lifecycle-phase">1. 실험적</div>
    <div class="lifecycle-prefix"><code>x-vendor-name</code></div>
    <div class="lifecycle-desc">벤더가 생성. 파괴적 변경 허용.</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">2. 스테이징</div>
    <div class="lifecycle-prefix"><code>s-name</code></div>
    <div class="lifecycle-desc">2개 이상의 렌더러 구현 필요. 거버넌스 검토.</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">3. 코어</div>
    <div class="lifecycle-prefix">접두사 없음</div>
    <div class="lifecycle-desc">ARP 사양의 일부. 추가적 변경만 허용.</div>
  </div>
</div>

</div>

<div class="blog-section">

## 적합성 수준

| 수준 | 필수 컴포넌트 | 대상 |
|-------|-------------------|--------|
| 최소 | `text` + 텍스트 입력 | 음성 어시스턴트, IoT |
| 기본 | text, table, form, confirm, choices | CLI 터미널 |
| 표준 | 모든 코어 컴포넌트, 전체 객체 모델 | 웹 / 데스크톱 |
| 전체 | 표준 + 스트리밍, 멀티 표면, 파일 업로드 | 리치 웹 앱 |

</div>

<div class="blog-cta">
  <a class="blog-cta-btn" href="/ko/docs/agentic/arp">전체 레퍼런스 읽기</a>
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
