/** Escape HTML special characters to prevent XSS. */
export function esc(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

/** Escape HTML including double quotes (for use in attributes). */
export function escAttr(s: string): string {
  return esc(s).replace(/"/g, "&quot;");
}

/**
 * Build a class string from a map of class names to booleans.
 * @example classMap({ active: true, hidden: false }) => "active"
 */
export function classMap(map: Record<string, boolean>): string {
  return Object.entries(map)
    .filter(([, v]) => v)
    .map(([k]) => k)
    .join(" ");
}

/**
 * Build an inline style string from a map of CSS property names to values.
 * Falsy values are omitted.
 * @example styleMap({ color: "red", display: "" }) => "color:red"
 */
export function styleMap(map: Record<string, string>): string {
  return Object.entries(map)
    .filter(([, v]) => v)
    .map(([k, v]) => `${k}:${v}`)
    .join(";");
}

/** Format a byte count as a human-readable string. */
export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/** Convert a tool name like "get_weather" to "Get Weather". */
export function formatToolName(name: string): string {
  return name.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}
