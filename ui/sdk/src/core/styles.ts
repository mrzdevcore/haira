import { css } from "lit";

// Design tokens — dark theme (default)
export const themeVars = css`
  --haira-bg: #09090b;
  --haira-bg-card: #0f0f12;
  --haira-bg-card-hover: #18181b;
  --haira-bg-elevated: #1c1c20;
  --haira-bg-input: #0c0c0f;
  --haira-border: rgba(63, 63, 70, 0.5);
  --haira-border-light: rgba(232, 163, 23, 0.12);
  --haira-border-focus: rgba(232, 163, 23, 0.4);
  --haira-accent: #e8a317;
  --haira-accent-light: #f0bd4f;
  --haira-accent-dim: rgba(232, 163, 23, 0.06);
  --haira-glow: #fde68a;
  --haira-text: #fafaf9;
  --haira-text-dim: #a1a1aa;
  --haira-muted: #71717a;
  --haira-success: #22c55e;
  --haira-error: #ef4444;
  --haira-warn: #eab308;
  --haira-info: #3b82f6;
  --haira-radius: 10px;
  --haira-radius-sm: 6px;
  --haira-font: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  --haira-mono: "SF Mono", "Fira Code", "JetBrains Mono", "Cascadia Code",
    monospace;
  --hljs-base: #d4d4d4;
  --hljs-keyword: #569cd6;
  --hljs-builtin: #dcdcaa;
  --hljs-type: #4ec9b0;
  --hljs-string: #ce9178;
  --hljs-comment: #6a9955;
  --hljs-number: #b5cea8;
  --hljs-regexp: #d16969;
  --hljs-symbol: #d7ba7d;
  --hljs-variable: #9cdcfe;
  --hljs-meta: #d7ba7d;
  --hljs-tag: #569cd6;
`;

// Light theme overrides — must cover ALL color variables from themeVars
export const lightThemeVarsStr = `
  --haira-bg: #ffffff;
  --haira-bg-card: #f7f7f8;
  --haira-bg-card-hover: #eeeff1;
  --haira-bg-elevated: #e8e8ec;
  --haira-bg-input: #f2f2f4;
  --haira-border: rgba(0, 0, 0, 0.1);
  --haira-border-light: rgba(0, 0, 0, 0.06);
  --haira-border-focus: rgba(0, 0, 0, 0.25);
  --haira-glow: #b8860b;
  --haira-text: #1a1a1a;
  --haira-text-dim: #4a4a4a;
  --haira-muted: #6b6b6b;
  --hljs-base: #000000;
  --hljs-keyword: #0000ff;
  --hljs-builtin: #795e26;
  --hljs-type: #267f99;
  --hljs-string: #a31515;
  --hljs-comment: #008000;
  --hljs-number: #098658;
  --hljs-regexp: #811f3f;
  --hljs-symbol: #e50000;
  --hljs-variable: #001080;
  --hljs-meta: #af00db;
  --hljs-tag: #800000;
`;

// Dark theme values as a string — used to restore dark theme from light
export const darkThemeVarsStr = `
  --haira-bg: #09090b;
  --haira-bg-card: #0f0f12;
  --haira-bg-card-hover: #18181b;
  --haira-bg-elevated: #1c1c20;
  --haira-bg-input: #0c0c0f;
  --haira-border: rgba(63, 63, 70, 0.5);
  --haira-border-light: rgba(232, 163, 23, 0.12);
  --haira-border-focus: rgba(232, 163, 23, 0.4);
  --haira-glow: #fde68a;
  --haira-text: #fafaf9;
  --haira-text-dim: #a1a1aa;
  --haira-muted: #71717a;
  --hljs-base: #d4d4d4;
  --hljs-keyword: #569cd6;
  --hljs-builtin: #dcdcaa;
  --hljs-type: #4ec9b0;
  --hljs-string: #ce9178;
  --hljs-comment: #6a9955;
  --hljs-number: #b5cea8;
  --hljs-regexp: #d16969;
  --hljs-symbol: #d7ba7d;
  --hljs-variable: #9cdcfe;
  --hljs-meta: #d7ba7d;
  --hljs-tag: #569cd6;
`;

// Keyframes
export const keyframes = css`
  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  @keyframes fadeSlideUp {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  @keyframes pop {
    0% {
      transform: scale(1);
    }
    50% {
      transform: scale(1.02);
    }
    100% {
      transform: scale(1);
    }
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.7;
    }
  }
  @keyframes blink {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.4;
    }
  }
  @keyframes bounce {
    0%,
    80%,
    100% {
      transform: translateY(0);
    }
    40% {
      transform: translateY(-6px);
    }
  }
  @keyframes expandDown {
    from {
      opacity: 0;
      max-height: 0;
    }
    to {
      opacity: 1;
      max-height: 600px;
    }
  }
`;

/** Base reset — inherits theme variables from haira-app root.
 *  Theme vars are only declared on <haira-app> :host so that
 *  light/dark switching propagates through all shadow DOMs. */
export const baseStyles = css`
  :host {
    font-family: var(--haira-font);
    color: var(--haira-text);
    -webkit-font-smoothing: antialiased;
  }
  *,
  *::before,
  *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }
  ${keyframes}
`;

/** Scrollbar styles */
export const scrollbarStyles = css`
  ::-webkit-scrollbar {
    width: 5px;
    height: 5px;
  }
  ::-webkit-scrollbar-track {
    background: transparent;
  }
  ::-webkit-scrollbar-thumb {
    background: var(--haira-muted);
    border-radius: 3px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background: var(--haira-accent);
  }
`;

/** Card surface */
export const cardStyles = css`
  .card {
    background: var(--haira-bg-card);
    border: 1px solid var(--haira-border);
    border-radius: var(--haira-radius);
    overflow: hidden;
  }
`;

/** Animate-in for generative UI components */
export const animateInStyles = css`
  :host {
    display: block;
    animation: fadeSlideUp 0.25s ease-out;
  }
`;

/** Shared highlight.js token styles — uses --hljs-* CSS custom properties from theme. */
export const hljsStyles = css`
  .hljs-keyword,
  .hljs-selector-tag,
  .hljs-deletion {
    color: var(--hljs-keyword);
  }
  .hljs-built_in,
  .hljs-title.function_,
  .hljs-title.function_.invoke__ {
    color: var(--hljs-builtin);
  }
  .hljs-title,
  .hljs-title.class_,
  .hljs-title.class_.inherited__ {
    color: var(--hljs-type);
  }
  .hljs-string,
  .hljs-template-tag,
  .hljs-template-variable {
    color: var(--hljs-string);
  }
  .hljs-comment,
  .hljs-quote {
    color: var(--hljs-comment);
    font-style: italic;
  }
  .hljs-number,
  .hljs-literal {
    color: var(--hljs-number);
  }
  .hljs-regexp {
    color: var(--hljs-regexp);
  }
  .hljs-symbol,
  .hljs-bullet {
    color: var(--hljs-symbol);
  }
  .hljs-attribute,
  .hljs-attr,
  .hljs-variable,
  .hljs-params {
    color: var(--hljs-variable);
  }
  .hljs-operator,
  .hljs-punctuation {
    color: var(--hljs-base);
  }
  .hljs-meta,
  .hljs-meta .hljs-keyword {
    color: var(--hljs-meta);
  }
  .hljs-section {
    color: var(--hljs-keyword);
    font-weight: bold;
  }
  .hljs-tag,
  .hljs-name {
    color: var(--hljs-tag);
  }
  .hljs-type {
    color: var(--hljs-type);
  }
  .hljs-addition {
    color: var(--hljs-comment);
  }
  .hljs-emphasis {
    font-style: italic;
  }
  .hljs-strong {
    font-weight: bold;
  }
`;

// Color helpers

export function methodColor(method: string): string {
  switch (method) {
    case "POST":
      return "#22c55e";
    case "GET":
      return "#3b82f6";
    case "PUT":
      return "#f59e0b";
    case "DELETE":
      return "#ef4444";
    default:
      return "#71717a";
  }
}

export function uiTypeColor(uiType: string): string {
  switch (uiType) {
    case "form":
      return "#3b82f6";
    case "chat":
      return "#22c55e";
    case "stream":
      return "#e8a317";
    default:
      return "#71717a";
  }
}

export function hexToRgb(
  hex: string
): { r: number; g: number; b: number } | null {
  const m = hex.match(/^#?([0-9a-f]{6})$/i);
  if (!m) return null;
  return {
    r: parseInt(m[1].substring(0, 2), 16),
    g: parseInt(m[1].substring(2, 4), 16),
    b: parseInt(m[1].substring(4, 6), 16),
  };
}

export function lighten(hex: string, amount: number): string {
  const rgb = hexToRgb(hex);
  if (!rgb) return hex;
  const r = Math.min(255, rgb.r + Math.round((255 - rgb.r) * amount));
  const g = Math.min(255, rgb.g + Math.round((255 - rgb.g) * amount));
  const b = Math.min(255, rgb.b + Math.round((255 - rgb.b) * amount));
  return `#${r.toString(16).padStart(2, "0")}${g.toString(16).padStart(2, "0")}${b.toString(16).padStart(2, "0")}`;
}
