import { lightThemeVarsStr, hexToRgb, lighten } from "../core/styles";

export function applyTheme(
  host: HTMLElement,
  options: { theme?: string; accent?: string }
) {
  if (options.theme === "light") {
    for (const line of lightThemeVarsStr.split("\n")) {
      const match = line.match(/(--[\w-]+):\s*(.+);/);
      if (match) host.style.setProperty(match[1], match[2].trim());
    }
  }

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
