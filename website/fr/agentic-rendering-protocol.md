---
layout: page
title: ARP — Protocole de Rendu Agentique
description: Un protocole indépendant du transport pour la communication entre agents IA et surfaces de rendu.
---

<div class="blog-page">
<div class="blog-hero">
  <div class="blog-hero-glow"></div>
  <div class="blog-hero-inner">
    <div class="blog-badge">Spécification du Protocole</div>
    <h1 class="blog-title">Protocole de Rendu Agentique</h1>
    <p class="blog-subtitle">Un protocole bidirectionnel indépendant du transport pour la communication entre agents IA et surfaces de rendu. Un seul agent, tous les écrans.</p>
    <div class="blog-meta">
      <div class="blog-meta-item">v0.1 Brouillon</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">Standard Ouvert</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">CC BY 4.0</div>
    </div>
  </div>
</div>

<div class="blog-body">

<div class="blog-section">

## Le Problème

Les agents IA produisent des sorties structurées — tableaux, formulaires, graphiques, messages de statut, blocs de code — mais le rendu de ces sorties est étroitement couplé au transport et au framework frontend. Un agent conçu pour un chat web ne peut pas s'afficher dans une application desktop. Un agent derrière un flux SSE ne peut pas servir un client mobile attendant un protocole différent.

Chaque nouvelle surface de rendu nécessite un code d'intégration personnalisé.

</div>

<div class="blog-section">

## La Solution

ARP définit un **protocole** — pas un framework, pas une bibliothèque. De la même façon que Wayland a découplé les applications des serveurs d'affichage, ARP découple les agents des renderers.

L'agent décide **quoi** afficher. Le renderer décide **comment** l'afficher.

<div class="blog-card">
<div class="blog-card-header">Architecture</div>

```
Agent (backend)                              Renderer (web, CLI, mobile)
───────────────                              ──────────────────────────
Owns surfaces         ── render / delta ──►  Owns display
Owns UI state         ◄── input events ────  Owns input routing
```

</div>
</div>

<div class="blog-section">

## Principes de Conception

<div class="principle-grid">
  <div class="principle-item">
    <div class="principle-num">01</div>
    <div class="principle-text">
      <div class="principle-title">Pas d'intermédiaire</div>
      <div class="principle-desc">L'agent envoie des commandes de rendu directement. Le renderer envoie les entrées directement. Pas de framework intermédiaire.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">02</div>
    <div class="principle-text">
      <div class="principle-title">Asynchrone &amp; non-bloquant</div>
      <div class="principle-desc">Tous les messages sont asynchrones. Aucun côté ne bloque en attendant une réponse.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">03</div>
    <div class="principle-text">
      <div class="principle-title">Chaque image est parfaite</div>
      <div class="principle-desc">Commits atomiques. L'état de rendu s'accumule dans un tampon en attente et s'applique atomiquement. Pas de scintillement.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">04</div>
    <div class="principle-text">
      <div class="principle-title">Piloté par les capacités</div>
      <div class="principle-desc">Les renderers déclarent ce qu'ils supportent. Les agents s'adaptent. Le CLI reçoit des tableaux, le web reçoit tout.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">05</div>
    <div class="principle-text">
      <div class="principle-title">Indépendant du transport</div>
      <div class="principle-desc">Messages logiques, pas formats filaires. WebSocket, SSE, gRPC, sockets Unix, stdio.</div>
    </div>
  </div>
  <div class="principle-item">
    <div class="principle-num">06</div>
    <div class="principle-text">
      <div class="principle-title">Composants typés</div>
      <div class="principle-desc">Les agents émettent des descripteurs typés comme <code>table</code> avec <code>headers</code> et <code>rows</code> — pas du HTML.</div>
    </div>
  </div>
</div>

</div>

<div class="blog-section">

## Messages du Protocole

Chaque message ARP est du JSON avec au minimum `{ v: 1, type: "<type>" }`.

<div class="blog-card">
<div class="blog-card-header">Serveur vers Client</div>

| Type | Utilité |
|------|---------|
| `hello` | Poignée de main des capacités à la connexion |
| `delta` | Fragment de texte incrémental |
| `tool_start` | Exécution d'outil démarrée |
| `tool_end` | Exécution d'outil terminée |
| `render` | Composant UI génératif |
| `patch` | Mise à jour incrémentale d'un composant |
| `error` | Événement d'erreur |
| `commit` | Flux terminé |

</div>

<div class="blog-card">
<div class="blog-card-header">Client vers Serveur</div>

| Type | Type d'Entrée | Utilité |
|------|--------------|---------|
| `input` | `text` | Message texte de l'utilisateur |
| `input` | `action` | Clic sur bouton / action UI |
| `input` | `form_submit` | Soumission de formulaire |

</div>
</div>

<div class="blog-section">

## 14 Composants Intégrés

Tout renderer conforme ARP doit supporter au minimum `text`. Les composants déclarent des chaînes de repli — `chart` revient à `table`, qui revient à `text`.

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

## Transports

<div class="transport-grid">
  <div class="transport-card transport-active">
    <div class="transport-label">Disponible</div>
    <div class="transport-name">WebSocket</div>
    <div class="transport-path"><code>/_arp/v1</code></div>
    <div class="transport-desc">Transport principal. Connexion bidirectionnelle persistante avec reconnexion automatique.</div>
  </div>
  <div class="transport-card transport-active">
    <div class="transport-label">Disponible</div>
    <div class="transport-name">SSE</div>
    <div class="transport-path">Server-Sent Events</div>
    <div class="transport-desc">Transport de repli. Requis pour les téléchargements de fichiers via multipart/form-data.</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">Prévu</div>
    <div class="transport-name">gRPC</div>
    <div class="transport-path">Haute performance</div>
    <div class="transport-desc">Pour les applications desktop et mobiles natives.</div>
  </div>
  <div class="transport-card">
    <div class="transport-label">Prévu</div>
    <div class="transport-name">stdio</div>
    <div class="transport-path">Tramage NDJSON</div>
    <div class="transport-desc">Pour les renderers CLI et les intégrations par tube.</div>
  </div>
</div>

</div>

<div class="blog-section">

## SDKs Client

<div class="blog-card">
<div class="blog-card-header">@haira/arp &mdash; Core (zéro dépendances)</div>

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
<div class="blog-card-header">@haira/arp-react &mdash; Interface Chat Prête à l'Emploi</div>

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

Également disponibles : **`@haira/arp-vue`** pour Vue 3 et **`github.com/haira-lang/arp-go`** pour les backends Go.

</div>

<div class="blog-section">

## Intégration Haira

Chaque serveur Haira parle ARP nativement. Aucune configuration requise.

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

## Cycle de Vie des Extensions

Les nouveaux composants suivent un cycle de vie en trois phases inspiré de Wayland :

<div class="lifecycle-steps">
  <div class="lifecycle-step">
    <div class="lifecycle-phase">1. Expérimental</div>
    <div class="lifecycle-prefix"><code>x-vendor-name</code></div>
    <div class="lifecycle-desc">Créé par le fournisseur. Modifications incompatibles autorisées.</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">2. Mise en scène</div>
    <div class="lifecycle-prefix"><code>s-name</code></div>
    <div class="lifecycle-desc">Nécessite 2+ implémentations de renderer. Revue de gouvernance.</div>
  </div>
  <div class="lifecycle-step">
    <div class="lifecycle-phase">3. Noyau</div>
    <div class="lifecycle-prefix">Sans préfixe</div>
    <div class="lifecycle-desc">Fait partie de la spécification ARP. Uniquement des modifications additives.</div>
  </div>
</div>

</div>

<div class="blog-section">

## Niveaux de Conformité

| Niveau | Composants Requis | Cible |
|--------|------------------|-------|
| Minimal | `text` + entrée texte | Assistants vocaux, IoT |
| Basique | text, table, form, confirm, choices | Terminaux CLI |
| Standard | Tous les composants noyau, modèle objet complet | Web / desktop |
| Complet | Standard + streaming, multi-surfaces, téléchargement de fichiers | Applications web riches |

</div>

<div class="blog-cta">
  <a class="blog-cta-btn" href="/fr/docs/agentic/arp">Lire la Référence Complète</a>
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
