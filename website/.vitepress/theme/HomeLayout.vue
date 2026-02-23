<script setup>
import { computed } from 'vue'
import { useData } from 'vitepress'
import DefaultTheme from 'vitepress/theme'

const { Layout } = DefaultTheme
const { page } = useData()

const isHome = computed(() => page.value.relativePath === 'index.md')
</script>

<template>
  <Layout>
    <!-- Gold "ai" in navbar logo on all pages -->
    <template #nav-bar-title-after>
      <span class="nav-haira-label">H<span class="haira-ai">ai</span>ra</span>
    </template>

    <!-- Landing page content rendered in the page content area -->
    <template v-if="isHome" #page-top>
      <div class="haira-landing">
        <div class="page-grid-pattern" />

        <!-- ======== HERO ======== -->
        <section class="hero">
          <div class="hero-glow" />
          <div class="hero-content">
            <div class="hero-badge">
              <span class="hero-badge-dot" />
              Open Source &middot; Apache 2.0
            </div>
            <h1 class="hero-title">
              The Programming Language for <span class="hero-accent">Agentic AI</span>
            </h1>
            <div class="hero-tagline">
              <span class="hero-tag">Four keywords</span>
              <span class="hero-tag-dot" />
              <span class="hero-tag">One binary</span>
              <span class="hero-tag-dot" />
              <span class="hero-tag">Zero boilerplate</span>
            </div>
            <p class="hero-subtitle">
              Build production AI agents with <strong>Generative UI</strong> &mdash; in minutes, not months.
            </p>
            <div class="hero-actions">
              <a href="/docs/getting-started/installation" class="hero-btn primary">
                Get Started
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M3 8h10m0 0L9 4m4 4L9 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </a>
              <a href="https://github.com/mrzdevcore/haira" target="_blank" class="hero-btn secondary">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
                Star on GitHub
              </a>
            </div>
            <div class="hero-install">
              <code>curl -fsSL https://haira.dev/install.sh | sh</code>
              <button class="hero-install-copy" onclick="navigator.clipboard.writeText('curl -fsSL https://haira.dev/install.sh | sh')" title="Copy">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><rect x="5" y="5" width="8" height="8" rx="1.5" stroke="currentColor" stroke-width="1.2"/><path d="M3 11V3.5A1.5 1.5 0 014.5 2H11" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
              </button>
            </div>
          </div>
        </section>

        <!-- ======== LOGO BAR ======== -->
        <section class="logo-bar">
          <div class="landing-container">
            <p class="logo-bar-label">Works with any LLM provider</p>
            <div class="logo-bar-logos">
              <span class="logo-bar-item">OpenAI</span>
              <span class="logo-bar-sep" />
              <span class="logo-bar-item">Anthropic</span>
              <span class="logo-bar-sep" />
              <span class="logo-bar-item">Azure OpenAI</span>
              <span class="logo-bar-sep" />
              <span class="logo-bar-item">Ollama</span>
              <span class="logo-bar-sep" />
              <span class="logo-bar-item">Groq</span>
              <span class="logo-bar-sep" />
              <span class="logo-bar-item">Mistral</span>
            </div>
          </div>
        </section>

        <!-- ======== CODE SHOWCASE ======== -->
        <section class="section code-section">
          <div class="landing-container">
            <div class="section-eyebrow">
              <span class="pill">See It In Action</span>
            </div>
            <h2 class="section-title">Four keywords.<br>That's the whole framework.</h2>
            <p class="section-desc">
              <code>provider</code> <code>tool</code> <code>agent</code> <code>workflow</code> &mdash; everything you need to build production AI agents.
            </p>

            <div class="showcase-grid">
              <div class="showcase-code">
                <div class="code-window">
                  <div class="code-window-bar">
                    <span class="dot r" /><span class="dot y" /><span class="dot g" />
                    <span class="code-filename">weather-agent.haira</span>
                  </div>
                  <div class="code-body">
<pre><code><span class="c-kw">import</span> <span class="c-str">"http"</span>

<span class="c-ag">provider</span> <span class="c-id">openai</span> {
    api_key: <span class="c-fn">env</span>(<span class="c-str">"OPENAI_API_KEY"</span>)
    model: <span class="c-str">"gpt-4o"</span>
}

<span class="c-ag">tool</span> <span class="c-id">get_weather</span>(city: <span class="c-tp">string</span>) -> <span class="c-tp">string</span> {
    <span class="c-doc">"""Get the current weather for a city."""</span>
    resp, err = http.get(<span class="c-str">"https://wttr.in/${city}?format=j1"</span>)
    <span class="c-kw">if</span> err != <span class="c-kw">nil</span> { <span class="c-kw">return</span> <span class="c-str">"Failed to fetch."</span> }
    data = resp.json()
    current = data[<span class="c-str">"current_condition"</span>][<span class="c-num">0</span>]
    <span class="c-kw">return</span> <span class="c-str">"${city}: ${current["temp_C"]}°C"</span>
}

<span class="c-ag">agent</span> <span class="c-id">Assistant</span> {
    provider: openai
    system: <span class="c-str">"You are a helpful assistant."</span>
    tools: [get_weather]
    memory: conversation(max_turns: <span class="c-num">10</span>)
}

<span class="c-dec">@post</span>(<span class="c-str">"/chat"</span>)
<span class="c-ag">workflow</span> <span class="c-id">Chat</span>(message: <span class="c-tp">string</span>) -> <span class="c-kw">stream</span> {
    <span class="c-kw">return</span> Assistant.stream(message)
}

<span class="c-kw">fn</span> <span class="c-id">main</span>() {
    http.Server([Chat]).listen(<span class="c-num">8080</span>)
}</code></pre>
                  </div>
                </div>
              </div>

              <div class="showcase-callouts">
                <div class="callout">
                  <div class="callout-num">1</div>
                  <div>
                    <h4>Provider</h4>
                    <p>Connect to any LLM &mdash; OpenAI, Anthropic, Azure, Ollama, or any OpenAI-compatible API. No SDK imports.</p>
                  </div>
                </div>
                <div class="callout">
                  <div class="callout-num">2</div>
                  <div>
                    <h4>Tool</h4>
                    <p>Give your agent capabilities. Type-checked functions with compiler-enforced documentation.</p>
                  </div>
                </div>
                <div class="callout">
                  <div class="callout-num">3</div>
                  <div>
                    <h4>Agent</h4>
                    <p>Declare an agent with tools, memory, and a system prompt. No class inheritance. No framework magic.</p>
                  </div>
                </div>
                <div class="callout">
                  <div class="callout-num">4</div>
                  <div>
                    <h4>Workflow</h4>
                    <p>Expose agents as HTTP endpoints. Streaming, forms, file uploads — all built in. Deploy as a single binary.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- ======== FEATURES ======== -->
        <section class="section features-section">
          <div class="landing-container wide">
            <div class="section-eyebrow">
              <span class="pill">Features</span>
            </div>
            <h2 class="section-title">Everything you need.<br>Nothing you don't.</h2>

            <div class="features-grid">
              <div class="feature-card">
                <div class="feature-icon">
                  <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>
                </div>
                <h3>Compiles to a Single Binary</h3>
                <p>Haira compiles to Go, then to a native binary. No interpreter, no VM, no <code>node_modules</code>. Ship one file, run anywhere.</p>
              </div>
              <div class="feature-card">
                <div class="feature-icon">
                  <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
                </div>
                <h3>Agentic by Design</h3>
                <p><code>provider</code>, <code>tool</code>, <code>agent</code>, <code>workflow</code> are first-class language keywords &mdash; not decorators bolted onto Python classes.</p>
              </div>
              <div class="feature-card">
                <div class="feature-icon">
                  <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/></svg>
                </div>
                <h3>Type-safe at Compile Time</h3>
                <p>Static types, pattern matching, enums, structs, pipe operator, and error handling. Catch bugs before they reach production.</p>
              </div>
              <div class="feature-card">
                <div class="feature-icon">
                  <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M15 10l-4 4l6 6l4-16l-18 7l4 2l2 6l3-4"/></svg>
                </div>
                <h3>Built-in Streaming & UI</h3>
                <p>SSE streaming with <code>-> stream</code>. Every workflow gets an auto-generated chat UI. Rich components via the ARP protocol.</p>
              </div>
              <div class="feature-card">
                <div class="feature-icon">
                  <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M8 7h12m0 0l-4-4m4 4l-4 4M16 17H4m0 0l4-4m-4 4l4 4"/></svg>
                </div>
                <h3>Multi-Agent Handoffs</h3>
                <p>Route between specialized agents automatically. Front desk to billing to tech &mdash; one config field, zero glue code.</p>
              </div>
              <div class="feature-card">
                <div class="feature-icon">
                  <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/></svg>
                </div>
                <h3>Batteries-Included Stdlib</h3>
                <p>HTTP, JSON, Postgres, Excel, Slack, GitHub, vector search, regex, time, file system &mdash; all built in, tree-shaken at compile time.</p>
              </div>
            </div>
          </div>
        </section>

        <!-- ======== GENERATIVE UI ======== -->
        <section class="section genui-section">
          <div class="landing-container">
            <div class="section-eyebrow">
              <span class="pill">Generative UI</span>
            </div>
            <h2 class="section-title">Your agent renders the interface.</h2>
            <p class="section-desc">Agents return rich UI components &mdash; tables, charts, status cards, confirmations &mdash; not just text. No frontend needed.</p>

            <div class="genui-grid">
              <!-- Code side -->
              <div class="genui-code">
                <div class="code-window">
                  <div class="code-window-bar">
                    <span class="dot r" /><span class="dot y" /><span class="dot g" />
                    <span class="code-filename">dashboard.haira</span>
                  </div>
                  <div class="code-body">
<pre><code><span class="c-kw">import</span> <span class="c-str">"ui"</span>

<span class="c-ag">tool</span> <span class="c-id">show_metrics</span>() -> <span class="c-tp">string</span> {
    <span class="c-doc">"""Show system metrics dashboard."""</span>
    ui.status_card(
        title: <span class="c-str">"API Health"</span>,
        value: <span class="c-str">"99.9%"</span>,
        status: <span class="c-str">"success"</span>
    )
    ui.table(
        headers: [<span class="c-str">"Service"</span>, <span class="c-str">"Status"</span>, <span class="c-str">"Latency"</span>],
        rows: [
            [<span class="c-str">"Auth"</span>, <span class="c-str">"Healthy"</span>, <span class="c-str">"12ms"</span>],
            [<span class="c-str">"DB"</span>, <span class="c-str">"Healthy"</span>, <span class="c-str">"3ms"</span>],
            [<span class="c-str">"Cache"</span>, <span class="c-str">"Warning"</span>, <span class="c-str">"89ms"</span>],
        ]
    )
    ui.chart(
        type: <span class="c-str">"line"</span>,
        title: <span class="c-str">"Requests / min"</span>,
        data: get_request_data()
    )
    <span class="c-kw">return</span> <span class="c-str">"Metrics displayed."</span>
}</code></pre>
                  </div>
                </div>
              </div>

              <!-- UI preview side -->
              <div class="genui-preview">
                <div class="genui-preview-window">
                  <div class="genui-preview-bar">
                    <span class="genui-preview-dot" />
                    <span class="genui-preview-title">Agent Response</span>
                  </div>
                  <div class="genui-preview-body">
                    <!-- Status card mock -->
                    <div class="mock-status-card">
                      <div class="mock-status-icon mock-success">
                        <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 8l3 3 5-5"/></svg>
                      </div>
                      <div>
                        <div class="mock-status-label">API Health</div>
                        <div class="mock-status-value">99.9% <span class="mock-status-badge">Healthy</span></div>
                      </div>
                    </div>

                    <!-- Table mock -->
                    <div class="mock-table">
                      <div class="mock-table-row mock-table-header">
                        <span>Service</span><span>Status</span><span>Latency</span>
                      </div>
                      <div class="mock-table-row">
                        <span>Auth</span><span class="mock-tag-green">Healthy</span><span>12ms</span>
                      </div>
                      <div class="mock-table-row">
                        <span>DB</span><span class="mock-tag-green">Healthy</span><span>3ms</span>
                      </div>
                      <div class="mock-table-row">
                        <span>Cache</span><span class="mock-tag-yellow">Warning</span><span>89ms</span>
                      </div>
                    </div>

                    <!-- Chart mock -->
                    <div class="mock-chart">
                      <div class="mock-chart-title">Requests / min</div>
                      <div class="mock-chart-bars">
                        <div class="mock-bar" style="height: 40%"></div>
                        <div class="mock-bar" style="height: 65%"></div>
                        <div class="mock-bar" style="height: 55%"></div>
                        <div class="mock-bar" style="height: 80%"></div>
                        <div class="mock-bar" style="height: 70%"></div>
                        <div class="mock-bar" style="height: 90%"></div>
                        <div class="mock-bar" style="height: 85%"></div>
                        <div class="mock-bar" style="height: 75%"></div>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="genui-components">
                  <span class="genui-chip">ui.status_card()</span>
                  <span class="genui-chip">ui.table()</span>
                  <span class="genui-chip">ui.chart()</span>
                  <span class="genui-chip">ui.key_value()</span>
                  <span class="genui-chip">ui.confirm()</span>
                  <span class="genui-chip">ui.product_cards()</span>
                  <span class="genui-chip">ui.group()</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- ======== COMPARISON ======== -->
        <section class="section compare-section">
          <div class="landing-container">
            <div class="section-eyebrow">
              <span class="pill">Why Haira?</span>
            </div>
            <h2 class="section-title">Replace your entire stack.</h2>
            <p class="section-desc">One language instead of a fragile tower of frameworks, SDKs, and YAML.</p>

            <div class="compare-grid">
              <div class="compare-card compare-before">
                <div class="compare-header">
                  <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M18 18L6 6M6 18L18 6"/></svg>
                  <span>The Old Way</span>
                </div>
                <ul>
                  <li>Python + LangChain + FastAPI</li>
                  <li>300+ lines of boilerplate per agent</li>
                  <li><code>pip install</code> 47 packages</li>
                  <li>Docker + requirements.txt + .env</li>
                  <li>Runtime type errors in production</li>
                  <li>Manual memory management</li>
                  <li>Build a separate frontend</li>
                </ul>
              </div>
              <div class="compare-card compare-after">
                <div class="compare-header">
                  <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M5 13l4 4L19 7"/></svg>
                  <span>With Haira</span>
                </div>
                <ul>
                  <li>One <code>.haira</code> file</li>
                  <li>30 lines for the same agent</li>
                  <li>Zero external dependencies</li>
                  <li>Single binary &mdash; copy and run</li>
                  <li>Compile-time type checking</li>
                  <li>Built-in conversation memory</li>
                  <li>Auto-generated UI included</li>
                </ul>
              </div>
            </div>
          </div>
        </section>

        <!-- ======== USE CASES ======== -->
        <section class="section usecases-section">
          <div class="landing-container wide">
            <div class="section-eyebrow">
              <span class="pill">Use Cases</span>
            </div>
            <h2 class="section-title">From chatbots to<br>enterprise workflows.</h2>

            <div class="usecases-grid">
              <div class="usecase">
                <h4>Internal Tools</h4>
                <p>Data explorers, admin dashboards, and ops tools with generative UI. Tables, charts, and status cards &mdash; rendered by the agent.</p>
              </div>
              <div class="usecase">
                <h4>Customer Support</h4>
                <p>Multi-agent handoffs, conversation memory, and session persistence. Route from triage to billing to tech &mdash; automatically.</p>
              </div>
              <div class="usecase">
                <h4>Workflow Automation</h4>
                <p>Replace n8n, Make, and Zapier with type-safe compiled workflows. Webhooks, triggers, and parallel execution built in.</p>
              </div>
              <div class="usecase">
                <h4>RAG Pipelines</h4>
                <p>Vector search, Postgres, and document processing in the standard library. Retrieval-augmented agents without glue code.</p>
              </div>
              <div class="usecase">
                <h4>DevOps & Incident Response</h4>
                <p>Agents that check health, query logs, create tickets, and notify teams &mdash; all from a single chat interface.</p>
              </div>
              <div class="usecase">
                <h4>Multi-Provider Agents</h4>
                <p>GPT-4o for reasoning, Claude for analysis, Ollama for private data &mdash; in the same app. Switch with one line.</p>
              </div>
            </div>
          </div>
        </section>

        <!-- ======== CTA ======== -->
        <section class="section cta-section">
          <div class="landing-container">
            <div class="cta-box">
              <div class="cta-glow" />
              <h2>Start building in seconds.</h2>
              <p>Install Haira, write your first agent, deploy a binary.</p>
              <div class="cta-actions">
                <a href="/docs/getting-started/installation" class="hero-btn primary">
                  Read the Docs
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M3 8h10m0 0L9 4m4 4L9 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                </a>
                <a href="https://github.com/mrzdevcore/haira" target="_blank" class="hero-btn secondary">
                  <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
                  Star on GitHub
                </a>
              </div>
              <div class="cta-github-nudge">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                If Haira helps you, consider giving it a star &mdash; it means a lot!
              </div>
            </div>
          </div>
        </section>

        <!-- ======== FOOTER ======== -->
        <footer class="landing-footer">
          <div class="landing-container">
            <div class="footer-inner">
              <div class="footer-brand">
                <span class="footer-logo">H<span class="haira-ai">ai</span>ra</span>
                <span class="footer-tagline">Build agents and workflows, not boilerplate.</span>
              </div>
              <div class="footer-links">
                <div class="footer-col">
                  <h5>Learn</h5>
                  <a href="/docs/getting-started/installation">Installation</a>
                  <a href="/docs/getting-started/hello-world">Hello World</a>
                  <a href="/docs/getting-started/key-concepts">Key Concepts</a>
                  <a href="/docs/examples">Examples</a>
                </div>
                <div class="footer-col">
                  <h5>Agentic</h5>
                  <a href="/docs/agentic/providers">Providers</a>
                  <a href="/docs/agentic/agents">Agents</a>
                  <a href="/docs/agentic/workflows">Workflows</a>
                  <a href="/docs/agentic/generative-ui">Generative UI</a>
                </div>
                <div class="footer-col">
                  <h5>Community</h5>
                  <a href="https://github.com/mrzdevcore/haira" target="_blank">GitHub</a>
                  <a href="https://github.com/mrzdevcore/haira/releases" target="_blank">Releases</a>
                  <a href="https://github.com/mrzdevcore/haira/issues" target="_blank">Issues</a>
                </div>
              </div>
            </div>
            <div class="footer-bottom">
              <span>By <a href="https://github.com/mrzdevcore" target="_blank">mrzdevcore</a> &middot; Released under the Apache-2.0 License</span>
            </div>
          </div>
        </footer>

      </div>
    </template>
  </Layout>
</template>
