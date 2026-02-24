---
layout: page
title: ARP — エージェントレンダリングプロトコル
description: AIエージェントとレンダリングサーフェス間の通信のためのトランスポート非依存プロトコル。
---

<div class="blog-page">
<div class="blog-hero">
  <div class="blog-hero-glow"></div>
  <div class="blog-hero-inner">
    <div class="blog-badge">プロトコル仕様</div>
    <h1 class="blog-title">エージェントレンダリングプロトコル</h1>
    <p class="blog-subtitle">AIエージェントとレンダリングサーフェス間の通信のための、トランスポート非依存の双方向プロトコル。ひとつのエージェント、あらゆる画面へ。</p>
    <div class="blog-meta">
      <div class="blog-meta-item">v0.1 ドラフト</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">オープンスタンダード</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">CC BY 4.0</div>
    </div>
  </div>
</div>

<div class="blog-body">

<div class="blog-section">

## 課題

AIエージェントはテーブル、フォーム、チャート、ステータスメッセージ、コードブロックといった構造化された出力を生成します。しかし、その出力のレンダリングはトランスポートとフロントエンドフレームワークに密結合されています。Webチャット向けに構築されたエージェントはデスクトップアプリでレンダリングできません。SSEストリームを使うエージェントは、別のプロトコルを期待するモバイルクライアントにサービスを提供できません。

新しいレンダリングサーフェスごとに、カスタムの統合コードが必要になります。

</div>

<div class="blog-section">

## 解決策

ARPは**プロトコル**を定義します — フレームワークでも、ライブラリでもありません。Waylandがアプリケーションをディスプレイサーバーからデカップリングしたように、ARPはエージェントをレンダラーからデカップリングします。

エージェントは**何を**表示するかを決定します。レンダラーは**どのように**表示するかを決定します。

<div class="blog-card">
<div class="blog-card-header">アーキテクチャ</div>

```
Agent (backend)                              Renderer (web, CLI, mobile)
───────────────                              ──────────────────────────
Owns surfaces         ── render / delta ──►  Owns display
Owns UI state         ◄── input events ────  Owns input routing
```

</div>
</div>

<div class="blog-section">

## 設計原則

<div class="principle-grid">
  <div class="principle-item">
    <div class="principle-num">01</div>
    <div class="principle-text">
      <div class="principle-title">中間者なし</div>
      <div class="principle-desc">エージェントはレンダリングコマンドを直接送信します。レンダラーは入力を直接送信します。中間フレームワークは不要です。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">02</div>
    <div class="principle-text">
      <div class="principle-title">非同期 &amp; ノンブロッキング</div>
      <div class="principle-desc">すべてのメッセージは非同期です。どちらの側も応答を待ってブロックしません。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">03</div>
    <div class="principle-text">
      <div class="principle-title">すべてのフレームが完全</div>
      <div class="principle-desc">アトミックコミット。レンダリング状態はペンディングバッファーに蓄積され、アトミックに適用されます。ちらつきはありません。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">04</div>
    <div class="principle-text">
      <div class="principle-title">ケイパビリティ駆動</div>
      <div class="principle-desc">レンダラーはサポート内容を宣言します。エージェントはそれに適応します。CLIはテーブルを取得し、Webはすべてを取得します。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">05</div>
    <div class="principle-text">
      <div class="principle-title">トランスポート非依存</div>
      <div class="principle-desc">論理メッセージであり、ワイヤーフォーマットではありません。WebSocket、SSE、gRPC、Unixソケット、stdio。</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">06</div>
    <div class="principle-text">
      <div class="principle-title">型付きコンポーネント</div>
      <div class="principle-desc">エージェントはHTMLではなく、<code>headers</code>と<code>rows</code>を持つ<code>table</code>のような型付きディスクリプタを送出します。</div>
    </div>
  </div>
</div>

</div>

<div class="blog-section">

## プロトコルメッセージ

すべてのARPメッセージは、最低限 `{ v: 1, type: "<type>" }` を持つJSONです。

<div class="blog-card">
<div class="blog-card-header">サーバーからクライアントへ</div>

| タイプ | 目的 |
|------|---------|
| `hello` | 接続時のケイパビリティハンドシェイク |
| `delta` | インクリメンタルなテキストチャンク |
| `tool_start` | ツール実行開始 |
| `tool_end` | ツール実行完了 |
| `render` | 生成的UIコンポーネント |
| `patch` | インクリメンタルなコンポーネント更新 |
| `error` | エラーイベント |
| `commit` | ストリーム完了 |

</div>

<div class="blog-card">
<div class="blog-card-header">クライアントからサーバーへ</div>

| タイプ | 入力タイプ | 目的 |
|------|-----------|---------|
| `input` | `text` | ユーザーのテキストメッセージ |
| `input` | `action` | ボタンクリック / UIアクション |
| `input` | `form_submit` | フォーム送信 |

</div>
</div>

<div class="blog-section">

## 14の組み込みコンポーネント

ARP準拠のレンダラーは最低限 `text` をサポートしなければなりません。コンポーネントはフォールバックチェーンを宣言します — `chart` は `table` にフォールバックし、`table` は `text` にフォールバックします。

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

## トランスポート

<div class="transport-grid">
  <div class="transport-card transport-active">
    <div class="transport-label">利用可能</div>
    <div class="transport-name">WebSocket</div>
    <div class="transport-path"><code>/_arp/v1</code></div>
    <div class="transport-desc">プライマリトランスポート。自動再接続付きの永続的な双方向接続。</div>
  </div>
  <div class="transport-card transport-active">
    <div class="transport-label">利用可能</div>
    <div class="transport-name">SSE</div>
    <div class="transport-path">Server-Sent Events</div>
    <div class="transport-desc">フォールバックトランスポート。multipart/form-dataによるファイルアップロードに必要。</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">計画中</div>
    <div class="transport-name">gRPC</div>
    <div class="transport-path">高パフォーマンス</div>
    <div class="transport-desc">ネイティブデスクトップおよびモバイルアプリケーション向け。</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">計画中</div>
    <div class="transport-name">stdio</div>
    <div class="transport-path">NDJSONフレーミング</div>
    <div class="transport-desc">CLIレンダラーおよびパイプベースの統合向け。</div>
  </div>
</div>

</div>

<div class="blog-section">

## クライアントSDK

<div class="blog-card">
<div class="blog-card-header">@haira/arp &mdash; コア（ゼロ依存）</div>

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
<div class="blog-card-header">@haira/arp-react &mdash; ドロップインチャットUI</div>

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

Vue 3向けの **`@haira/arp-vue`** と、Goバックエンド向けの **`github.com/haira-lang/arp-go`** も提供しています。

</div>

<div class="blog-section">

## Haira統合

すべてのHairaサーバーはネイティブでARPを話します。設定は不要です。

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

## 拡張ライフサイクル

新しいコンポーネントは、Waylandにインスパイアされた3フェーズのライフサイクルに従います：

<div class="lifecycle-steps">
  <div class="lifecycle-step">
    <div class="lifecycle-phase">1. 実験的</div>
    <div class="lifecycle-prefix"><code>x-vendor-name</code></div>
    <div class="lifecycle-desc">ベンダーが作成。破壊的変更を許容。</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">2. ステージング</div>
    <div class="lifecycle-prefix"><code>s-name</code></div>
    <div class="lifecycle-desc">2つ以上のレンダラー実装が必要。ガバナンスレビューあり。</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">3. コア</div>
    <div class="lifecycle-prefix">プレフィックスなし</div>
    <div class="lifecycle-desc">ARPスペックの一部。追加的な変更のみ。</div>
  </div>
</div>

</div>

<div class="blog-section">

## 適合レベル

| レベル | 必須コンポーネント | ターゲット |
|-------|-------------------|--------|
| 最小 | `text` + テキスト入力 | 音声アシスタント、IoT |
| 基本 | text、table、form、confirm、choices | CLIターミナル |
| 標準 | すべてのコアコンポーネント、完全なオブジェクトモデル | Web / デスクトップ |
| フル | 標準 + ストリーミング、マルチサーフェス、ファイルアップロード | リッチWebアプリ |

</div>

<div class="blog-cta">
  <a class="blog-cta-btn" href="/ja/docs/agentic/arp">完全なリファレンスを読む</a>
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
