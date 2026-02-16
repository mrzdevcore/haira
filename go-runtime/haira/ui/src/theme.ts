export const themeVars = `
  --haira-bg: #09090b;
  --haira-bg-card: #0f0f12;
  --haira-bg-card-hover: #18181b;
  --haira-bg-elevated: #1c1c20;
  --haira-bg-input: #0c0c0f;
  --haira-border: rgba(63, 63, 70, 0.5);
  --haira-border-light: rgba(232, 163, 23, 0.12);
  --haira-border-focus: rgba(232, 163, 23, 0.4);
  --haira-gold: #e8a317;
  --haira-gold-light: #f0bd4f;
  --haira-gold-dim: rgba(232, 163, 23, 0.06);
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
  --haira-font: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --haira-mono: 'SF Mono', 'Fira Code', 'JetBrains Mono', 'Cascadia Code', monospace;
`;

export const baseStyles = `
  :host {
    ${themeVars}
    font-family: var(--haira-font);
    color: var(--haira-text);
    -webkit-font-smoothing: antialiased;
  }
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
`;

export const scrollbarStyles = `
  ::-webkit-scrollbar { width: 5px; height: 5px; }
  ::-webkit-scrollbar-track { background: transparent; }
  ::-webkit-scrollbar-thumb { background: var(--haira-muted); border-radius: 3px; }
  ::-webkit-scrollbar-thumb:hover { background: var(--haira-gold); }
  scrollbar-width: thin;
  scrollbar-color: var(--haira-muted) transparent;
`;

export const sharedKeyframes = `
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  @keyframes fadeSlideUp {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }
  @keyframes pop {
    0% { transform: scale(1); }
    50% { transform: scale(1.02); }
    100% { transform: scale(1); }
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
  }
  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }
  @keyframes bounce {
    0%, 80%, 100% { transform: translateY(0); }
    40% { transform: translateY(-6px); }
  }
  @keyframes expandDown {
    from { opacity: 0; max-height: 0; }
    to { opacity: 1; max-height: 600px; }
  }
`;

// --- SVG Icons (inline strings) ---

export const iconPending = `<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6.5" stroke="currentColor" stroke-width="1.5" stroke-dasharray="3 2"/></svg>`;

export const iconSpinner = `<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-dasharray="28 10" style="animation:spin 0.7s linear infinite;transform-origin:center"/></svg>`;

export const iconCheck = `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M4.5 8.5L7 11L11.5 5.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`;

export const iconX = `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M5 5L11 11M11 5L5 11" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>`;

export const iconRetry = `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M3 8a5 5 0 0 1 8.5-3.5M13 8a5 5 0 0 1-8.5 3.5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/><path d="M11 2v3h-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M5 14v-3h3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>`;

export const iconCopy = `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><rect x="5" y="5" width="9" height="9" rx="1.5" stroke="currentColor" stroke-width="1.5"/><path d="M11 5V3.5A1.5 1.5 0 0 0 9.5 2H3.5A1.5 1.5 0 0 0 2 3.5v6A1.5 1.5 0 0 0 3.5 11H5" stroke="currentColor" stroke-width="1.5"/></svg>`;

export const iconCopyDone = `<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M4 8.5L6.5 11L12 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`;

export const iconChevron = `<svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`;

// Simplified 24px Haira robot head for header
export const logoSvg = `<svg width="22" height="22" viewBox="0 0 64 52" fill="none" xmlns="http://www.w3.org/2000/svg"><rect x="17" y="11" width="30" height="20" rx="6" fill="#F0BD4F"/><rect x="21" y="15" width="22" height="9" rx="4" fill="#3D2B1F"/><circle cx="27" cy="19.5" r="3.5" fill="#FDE68A"/><circle cx="27" cy="19.5" r="1.5" fill="#fff"/><circle cx="37" cy="19.5" r="3.5" fill="#FDE68A"/><circle cx="37" cy="19.5" r="1.5" fill="#fff"/><ellipse cx="32" cy="12" rx="25" ry="4" fill="#C4A265"/><ellipse cx="32" cy="11.5" rx="23" ry="3" fill="#D4B87A"/><rect x="18" y="2" width="28" height="10" rx="5" fill="#C4A265"/><rect x="20" y="1" width="24" height="5" rx="3" fill="#D4B87A"/><rect x="18" y="8" width="28" height="3.5" rx="1.5" fill="#5C3A1E"/><rect x="20" y="31" width="24" height="14" rx="4" fill="#E8A317"/><rect x="25" y="34" width="14" height="8" rx="3" fill="#3D2B1F"/><circle cx="32" cy="38" r="3" fill="#FDE68A"/><rect x="10" y="34" width="10" height="4" rx="2" fill="#E8A317"/><rect x="44" y="34" width="10" height="4" rx="2" fill="#E8A317"/></svg>`;

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
