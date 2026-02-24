---
layout: page
title: UI Générative
description: Des agents qui affichent des composants riches et interactifs — pas seulement du texte.
---

<div class="blog-page">
<div class="blog-hero">
  <div class="blog-hero-glow"></div>
  <div class="blog-hero-inner">
    <div class="blog-badge">Fonctionnalité du Langage</div>
    <h1 class="blog-title">UI Générative</h1>
    <p class="blog-subtitle">Des agents qui affichent des composants riches et interactifs en ligne — tableaux, graphiques, formulaires, diffs — pas seulement du texte. Aucun code frontend requis.</p>
    <div class="blog-meta">
      <div class="blog-meta-item">Intégré</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">11 Composants</div>
      <div class="blog-meta-sep"></div>
      <div class="blog-meta-item">Extensible</div>
    </div>
  </div>
</div>

<div class="blog-body">

<div class="blog-section">

## Au-delà du Texte

Les interfaces de chatbot traditionnelles affichent tout sous forme de bulles de texte. Mais les agents produisent des données structurées — résultats de requêtes, rapports de validation, statuts de déploiement, diffs de comparaison. Afficher tout cela en markdown est une mauvaise expérience.

L'UI Générative de Haira permet aux outils de contrôler déclarativement la façon dont leur sortie s'affiche. Une requête de base de données s'affiche en tableau. Un résultat de validation s'affiche en carte de statut. Un pipeline de déploiement s'affiche en suivi de progression. Tout cela en ligne dans le chat, sans écrire de code frontend.

</div>

<div class="blog-section">

## Comment Ça Fonctionne

Les outils retournent des composants UI en utilisant le module `ui`. Le runtime fait deux choses avec chaque résultat :

<div class="dual-grid">
  <div class="dual-card">
    <div class="dual-label">Pour l'Utilisateur</div>
    <div class="dual-title">Composant Riche</div>
    <div class="dual-desc">Diffusé vers le frontend via le <a href="/fr/agentic-rendering-protocol">protocole ARP</a> et affiché sous forme de tableau, carte, graphique ou formulaire interactif.</div>
  </div>
  <div class="dual-card">
    <div class="dual-label">Pour l'Agent</div>
    <div class="dual-title">Résumé Textuel</div>
    <div class="dual-desc">Une représentation textuelle compacte est envoyée au LLM pour qu'il puisse raisonner sur les données et décider des prochaines étapes.</div>
  </div>
</div>

Aucun chemin n'est sacrifié. L'utilisateur voit un beau tableau, l'agent voit des données structurées.

<div class="blog-card">
<div class="blog-card-header">Exemple &mdash; Outil avec Sortie UI</div>

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
```

</div>

Activez l'UI sur un agent avec une seule ligne :

```haira
agent DataExplorer {
    provider: OpenAI
    tools: [query_database, search_data]
    ui: ui
    memory: conversation(max_turns: 30)
}
```

</div>

<div class="blog-section">

## 11 Composants Intégrés

<div class="component-grid">
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
    </div>
    <div class="component-name">status-card</div>
    <div class="component-desc">Cartes de résultat succès, erreur, avertissement avec sections réductibles</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/></svg>
    </div>
    <div class="component-name">table</div>
    <div class="component-desc">Tableaux de données consultables et défilables avec onglets et en-têtes fixes</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
    </div>
    <div class="component-name">code-block</div>
    <div class="component-desc">Code avec coloration syntaxique, bouton de copie et onglets multiples</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="2" x2="12" y2="22"/><rect x="2" y="4" width="8" height="16" rx="1"/><rect x="14" y="4" width="8" height="16" rx="1"/></svg>
    </div>
    <div class="component-name">diff</div>
    <div class="component-desc">Comparaison avant/après côte à côte avec coloration syntaxique</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
    </div>
    <div class="component-name">key-value</div>
    <div class="component-desc">Listes de propriétés avec valeurs stylisées pour l'affichage de métadonnées</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
    </div>
    <div class="component-name">progress</div>
    <div class="component-desc">Suivi de pipeline multi-étapes avec statut par étape</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
    </div>
    <div class="component-name">chart</div>
    <div class="component-desc">Visualisations de données en courbes, barres, camembert, aires et nuages de points</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="8" y1="12" x2="16" y2="12"/><line x1="8" y1="16" x2="13" y2="16"/><line x1="8" y1="8" x2="16" y2="8"/></svg>
    </div>
    <div class="component-name">form</div>
    <div class="component-desc">Formulaires interactifs avec champs texte, sélection, case à cocher et zone de texte</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
    </div>
    <div class="component-name">confirm</div>
    <div class="component-desc">Dialogues de confirmation oui/non pour les actions destructives</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/><circle cx="5" cy="12" r="1"/></svg>
    </div>
    <div class="component-name">choices</div>
    <div class="component-desc">Sélecteurs d'options par bouton ou liste pour le choix de l'utilisateur</div>
  </div>
  <div class="component-card">
    <div class="component-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
    </div>
    <div class="component-name">product-cards</div>
    <div class="component-desc">Grilles de cartes image pour le e-commerce et les affichages de catalogue</div>
  </div>
</div>

<div class="blog-card">
<div class="blog-card-header">Utilisation</div>

```haira
import "ui"

ui.status_card("success", "Deploy Complete", "All 3 services updated")
ui.table("Results", ["Name", "Email"], [["Alice", "a@co"], ["Bob", "b@co"]])
ui.key_value("Server Info", {"Region": "us-east-1", "Status": "healthy"})
ui.chart("bar", "Revenue", ["Q1", "Q2", "Q3", "Q4"], [dataset])
ui.confirm("Delete this record?", "Yes, delete", "Cancel")
ui.group(
    ui.status_card("success", "Query Complete", "42 rows"),
    ui.table("Results", headers, rows)
)
```

</div>
</div>

<div class="blog-section">

## Interface Web Auto-générée

Chaque workflow Haira obtient automatiquement une interface web. Définissez un workflow, Haira génère le formulaire :

```haira
@webui(title: "File Summarizer", description: "Upload a file and get an AI summary")
@post("/summarize")
workflow Summarize(document: file, context: string) -> { summary: string } {
    content, err = io.read_file(document)
    if err != nil { return { summary: "Failed to read file." } }
    reply, err = Summarizer.ask("Summarize: ${content}")
    if err != nil { return { summary: "AI error." } }
    return { summary: reply }
}
```

Le décorateur `@webui` définit le titre et la description. Les paramètres `file` s'affichent en champs d'envoi. Les workflows en streaming (`-> stream`) obtiennent une interface de chat complète.

</div>

<div class="blog-section">

## Interface Chat en Streaming

Les workflows en streaming bénéficient de l'expérience la plus riche — streaming de tokens en temps réel, cartes d'exécution d'outils et composants UI en ligne :

```haira
@webhook("/chat")
workflow Chat(message: string, session_id: string) -> stream {
    return Assistant.stream(message, session: session_id)
}

fn main() {
    http.Server([Chat]).listen(8080)
}
```

L'interface de chat communique via le [protocole ARP](/fr/agentic-rendering-protocol) — gérant les deltas de texte, les événements du cycle de vie des outils et le rendu de composants riches via WebSocket ou SSE.

</div>

<div class="blog-section">

## Composants Personnalisés

Pour des besoins spécifiques à un domaine, déposez des Web Components TypeScript dans un répertoire `components/` :

<div class="blog-card">
<div class="blog-card-header">components/gantt-chart.ts</div>

```typescript
export class HairaGanttChart extends HTMLElement {
  connectedCallback() {
    this.attachShadow({ mode: "open" });
  }
  setProps(props) {
    // Render your custom UI
  }
}

export default {
  tag: "haira-gantt-chart",
  component: HairaGanttChart,
};
```

</div>

Les composants personnalisés héritent du thème Haira via des propriétés CSS personnalisées et peuvent dispatcher des événements `haira-action` qui deviennent des messages de chat. Le compilateur les découvre, les regroupe et les intègre au moment de la compilation.

</div>

<div class="blog-section">

## Le Pipeline

<div class="pipeline-grid">
  <div class="pipeline-step">
    <div class="pipeline-num">1</div>
    <div class="pipeline-label">Outils</div>
    <div class="pipeline-detail">Retournent des données typées via les fonctions <code>ui.*</code></div>
  </div>
  <div class="pipeline-arrow">
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14m0 0l-4-4m4 4l-4 4"/></svg>
  </div>
  <div class="pipeline-step">
    <div class="pipeline-num">2</div>
    <div class="pipeline-label">Runtime</div>
    <div class="pipeline-detail">Émet des messages ARP via WebSocket ou SSE</div>
  </div>
  <div class="pipeline-arrow">
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14m0 0l-4-4m4 4l-4 4"/></svg>
  </div>
  <div class="pipeline-step">
    <div class="pipeline-num">3</div>
    <div class="pipeline-label">Frontend</div>
    <div class="pipeline-detail">Affiche le composant correspondant en ligne dans le chat</div>
  </div>
</div>

Pas de dépôt frontend séparé. Pas de client API à maintenir. Pas de bibliothèque de composants à installer. Un seul fichier `.haira`, un seul binaire, une interface complète.

</div>

<div class="blog-cta">
  <a class="blog-cta-btn" href="/fr/docs/agentic/generative-ui">Lire la Référence Complète</a>
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
  max-width: 800px;
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
  font-size: 4rem;
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1.08;
  background: linear-gradient(135deg, var(--gold) 0%, var(--gold-light) 50%, var(--gold-glow) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 1.5rem;
  padding: 0 1rem;
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
.blog-card div[class*="language-"] {
  margin: 0 !important;
  border-radius: 0 !important;
}

/* ── Dual Grid ── */
.dual-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin: 1.5rem 0;
}
.dual-card {
  padding: 1.75rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background: var(--vp-c-bg-soft);
}
.dual-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--gold);
  margin-bottom: 0.5rem;
}
.dual-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin-bottom: 0.625rem;
}
.dual-desc {
  font-size: 0.9375rem;
  color: var(--vp-c-text-3);
  line-height: 1.6;
}
.dual-desc a {
  color: var(--gold);
  text-decoration: none;
}

/* ── Component Grid ── */
.component-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 0.75rem;
  margin: 1.5rem 0;
}
.component-card {
  display: flex;
  gap: 1rem;
  padding: 1.25rem 1.375rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background: var(--vp-c-bg-soft);
  align-items: flex-start;
  transition: border-color 0.15s;
}
.component-card:hover {
  border-color: rgba(232, 163, 23, 0.3);
}
.component-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  background: rgba(232, 163, 23, 0.08);
  color: var(--gold);
}
.component-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
  font-family: 'JetBrains Mono', monospace;
  margin-bottom: 0.25rem;
}
.component-desc {
  font-size: 0.8125rem;
  color: var(--vp-c-text-3);
  line-height: 1.5;
}

/* ── Pipeline ── */
.pipeline-grid {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  margin: 1.75rem 0;
}
.pipeline-step {
  flex: 1;
  padding: 1.5rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background: var(--vp-c-bg-soft);
  text-align: center;
}
.pipeline-num {
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--gold);
  font-family: 'JetBrains Mono', monospace;
  margin-bottom: 0.5rem;
}
.pipeline-label {
  font-size: 1.0625rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin-bottom: 0.375rem;
}
.pipeline-detail {
  font-size: 0.8125rem;
  color: var(--vp-c-text-3);
  line-height: 1.5;
}
.pipeline-arrow {
  flex-shrink: 0;
  color: var(--vp-c-text-3);
  opacity: 0.4;
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

/* ── Responsive ── */
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
  .blog-title { font-size: 2.5rem; }
  .blog-subtitle { font-size: 1.0625rem; }
  .blog-body { padding: 0 1.25rem 4rem; }
  .blog-body p { font-size: 1rem; }
  .dual-grid { grid-template-columns: 1fr; }
  .component-grid { grid-template-columns: 1fr; }
  .pipeline-grid { flex-direction: column; }
  .pipeline-arrow { transform: rotate(90deg); }
}
</style>
