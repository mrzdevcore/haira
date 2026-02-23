import type { StatusCardProps } from "@haira/arp";

const STATUS_COLORS: Record<string, string> = {
  success: "#22c55e",
  error: "#ef4444",
  warning: "#f59e0b",
  info: "#3b82f6",
};

const STATUS_ICONS: Record<string, string> = {
  success: "✓",
  error: "✗",
  warning: "⚠",
  info: "ℹ",
};

export function StatusCard({
  status,
  title,
  message,
  sections,
}: StatusCardProps & { onInput?: (text: string) => void }) {
  const color = STATUS_COLORS[status] ?? STATUS_COLORS.info;

  return (
    <div
      className="arp-status-card"
      style={{
        borderLeft: `3px solid ${color}`,
        background: "#111118",
        borderRadius: 8,
        padding: "12px 16px",
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <span style={{ color, fontSize: 16 }}>{STATUS_ICONS[status] ?? "ℹ"}</span>
        <span style={{ fontWeight: 600, color: "#e0e0e0", fontSize: 14 }}>{title}</span>
      </div>
      {message && (
        <div style={{ marginTop: 8, color: "#a0a0b0", fontSize: 13, lineHeight: 1.5 }}>
          {message}
        </div>
      )}
      {sections && sections.length > 0 && (
        <div style={{ marginTop: 10, display: "flex", flexDirection: "column", gap: 6 }}>
          {sections.map((s, i) => (
            <div key={i}>
              <div style={{ fontSize: 11, color: "#666", textTransform: "uppercase", letterSpacing: 0.5 }}>
                {s.label}
              </div>
              <div style={{ fontSize: 13, color: s.style ? (STATUS_COLORS[s.style] ?? "#e0e0e0") : "#e0e0e0" }}>
                {s.content}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
