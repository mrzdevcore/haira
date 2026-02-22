// Theme manager — applies light/dark theme and accent color overrides.
// Extracted from haira-app.ts for reuse.

import { lightThemeVars, hexToRgb, lighten } from "../core/styles";

/**
 * Apply theme overrides to a host element's CSS custom properties.
 * Call this on the shadow host to set light mode and/or custom accent color.
 */
export function applyTheme(
  host: HTMLElement,
  options: { theme?: string; accent?: string },
) {
  // Light theme overrides
  if (options.theme === "light") {
    for (const line of lightThemeVars.split("\n")) {
      const match = line.match(/(--[\w-]+):\s*(.+);/);
      if (match) host.style.setProperty(match[1], match[2].trim());
    }
  }

  // Accent color overrides
  const accent = options.accent;
  if (accent) {
    host.style.setProperty("--haira-accent", accent);
    host.style.setProperty("--haira-accent-light", lighten(accent, 0.25));
    const rgb = hexToRgb(accent);
    if (rgb) {
      host.style.setProperty("--haira-accent-dim", `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.06)`);
      host.style.setProperty("--haira-border-light", `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.12)`);
      host.style.setProperty("--haira-border-focus", `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.4)`);
    }
  }
}
