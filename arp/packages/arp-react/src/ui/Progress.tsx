import type { ProgressProps } from "@haira/arp";

const STATUS_STYLES: Record<string, { color: string; icon: string }> = {
  done: { color: "#22c55e", icon: "✓" },
  completed: { color: "#22c55e", icon: "✓" },
  running: { color: "#e8a317", icon: "●" },
  active: { color: "#e8a317", icon: "●" },
  failed: { color: "#ef4444", icon: "✗" },
  error: { color: "#ef4444", icon: "✗" },
  pending: { color: "#555", icon: "○" },
};

export function Progress({
  title,
  steps,
}: ProgressProps & { onInput?: (text: string) => void }) {
  return (
    <div className="arp-progress" style={{ background: "#111118", borderRadius: 8, padding: "12px 16px" }}>
      {title && (
        <div style={{ fontWeight: 600, fontSize: 14, color: "#e0e0e0", marginBottom: 12 }}>
          {title}
        </div>
      )}
      <div style={{ display: "flex", flexDirection: "column", gap: 0 }}>
        {steps.map((step, i) => {
          const s = STATUS_STYLES[step.status] ?? STATUS_STYLES.pending;
          return (
            <div key={i} style={{ display: "flex", gap: 12, alignItems: "flex-start" }}>
              {/* Connector */}
              <div style={{ display: "flex", flexDirection: "column", alignItems: "center", width: 20 }}>
                <span style={{ color: s.color, fontSize: 14, lineHeight: "24px" }}>{s.icon}</span>
                {i < steps.length - 1 && (
                  <div style={{ width: 1, height: 20, background: "#2a2a3e" }} />
                )}
              </div>
              {/* Content */}
              <div style={{ paddingBottom: i < steps.length - 1 ? 8 : 0 }}>
                <div style={{ fontSize: 13, color: "#e0e0e0" }}>{step.name}</div>
                {step.detail && (
                  <div style={{ fontSize: 12, color: "#666", marginTop: 2 }}>{step.detail}</div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
