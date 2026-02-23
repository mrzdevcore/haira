import type { DiffProps } from "@haira/arp";

export function Diff({
  title,
  before_label = "Before",
  after_label = "After",
  before,
  after,
}: DiffProps & { onInput?: (text: string) => void }) {
  return (
    <div className="arp-diff" style={{ background: "#111118", borderRadius: 8, overflow: "hidden" }}>
      {title && (
        <div style={{ padding: "10px 16px", borderBottom: "1px solid #1a1a2e", fontWeight: 600, fontSize: 14, color: "#e0e0e0" }}>
          {title}
        </div>
      )}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 0 }}>
        <div style={{ borderRight: "1px solid #1a1a2e" }}>
          <div style={{ padding: "6px 12px", fontSize: 11, color: "#ef4444", fontWeight: 600, borderBottom: "1px solid #1a1a2e" }}>
            {before_label}
          </div>
          <pre style={{ margin: 0, padding: "12px 16px", overflow: "auto", fontSize: 13, lineHeight: 1.5, color: "#f87171", background: "rgba(239,68,68,0.04)" }}>
            {before}
          </pre>
        </div>
        <div>
          <div style={{ padding: "6px 12px", fontSize: 11, color: "#22c55e", fontWeight: 600, borderBottom: "1px solid #1a1a2e" }}>
            {after_label}
          </div>
          <pre style={{ margin: 0, padding: "12px 16px", overflow: "auto", fontSize: 13, lineHeight: 1.5, color: "#4ade80", background: "rgba(34,197,94,0.04)" }}>
            {after}
          </pre>
        </div>
      </div>
    </div>
  );
}
