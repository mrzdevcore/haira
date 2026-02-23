import type { KeyValueProps } from "@haira/arp";

const STYLE_COLORS: Record<string, string> = {
  success: "#22c55e",
  error: "#ef4444",
  warning: "#f59e0b",
  info: "#3b82f6",
  code: "#a78bfa",
};

export function KeyValue({
  title,
  items,
}: KeyValueProps & { onInput?: (text: string) => void }) {
  return (
    <div className="arp-key-value" style={{ background: "#111118", borderRadius: 8, overflow: "hidden" }}>
      {title && (
        <div style={{ padding: "10px 16px", borderBottom: "1px solid #1a1a2e", fontWeight: 600, fontSize: 14, color: "#e0e0e0" }}>
          {title}
        </div>
      )}
      <div style={{ padding: "4px 0" }}>
        {items.map((item, i) => (
          <div
            key={i}
            style={{
              display: "flex",
              padding: "8px 16px",
              gap: 16,
              fontSize: 13,
              borderBottom: i < items.length - 1 ? "1px solid #0a0a12" : undefined,
            }}
          >
            <span style={{ color: "#888", minWidth: 120, flexShrink: 0 }}>{item.key}</span>
            <span style={{ color: item.style ? (STYLE_COLORS[item.style] ?? "#e0e0e0") : "#e0e0e0", fontFamily: item.style === "code" ? "monospace" : "inherit" }}>
              {item.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
