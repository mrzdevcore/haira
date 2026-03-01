import {
  lightThemeVarsStr,
  darkThemeVarsStr,
  hexToRgb,
  lighten,
} from "../core/styles";

function applyVarsStr(host: HTMLElement, varsStr: string) {
  for (const line of varsStr.split("\n")) {
    const match = line.match(/(--[\w-]+):\s*(.+);/);
    if (match) host.style.setProperty(match[1], match[2].trim());
  }
}

export function applyTheme(
  host: HTMLElement,
  options: { theme?: string; accent?: string }
) {
  // Apply base theme colors
  if (options.theme === "light") {
    applyVarsStr(host, lightThemeVarsStr);
  } else {
    // Dark theme — restore defaults (needed when switching back from light)
    applyVarsStr(host, darkThemeVarsStr);
  }

  // Apply accent color overrides
  const accent = options.accent;
  if (accent) {
    host.style.setProperty("--haira-accent", accent);
    host.style.setProperty("--haira-accent-light", lighten(accent, 0.25));
    const rgb = hexToRgb(accent);
    if (rgb) {
      host.style.setProperty(
        "--haira-accent-dim",
        `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.06)`
      );
      // Only override border vars for dark theme (light theme has its own)
      if (options.theme !== "light") {
        host.style.setProperty(
          "--haira-border-light",
          `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.12)`
        );
        host.style.setProperty(
          "--haira-border-focus",
          `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.4)`
        );
      }
    }
  }
}
