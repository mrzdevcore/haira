import type { ToolCardState } from "../hooks/use-arp-chat.js";

// ---------------------------------------------------------------------------
// Props
// ---------------------------------------------------------------------------

export interface ArpActivityPanelProps {
  /** Active tool executions. */
  toolCards: ToolCardState[];
  /** Whether the panel is visible. */
  open: boolean;
  /** Toggle panel visibility. */
  onToggle?: () => void;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

/**
 * Displays active tool executions with status indicators and elapsed time.
 */
export function ArpActivityPanel({
  toolCards,
  open,
  onToggle,
}: ArpActivityPanelProps) {
  if (toolCards.length === 0) return null;

  const runningCount = toolCards.filter((t) => t.status === "running").length;

  return (
    <div
      className="arp-activity-panel"
      style={{
        borderLeft: "1px solid #2a2a3e",
        background: "#111118",
        overflow: "hidden",
        transition: "width 0.2s ease",
        width: open ? 280 : 0,
      }}
    >
      {open && (
        <div style={{ padding: 16 }}>
          {/* Header */}
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              marginBottom: 12,
            }}
          >
            <span style={{ color: "#a0a0b0", fontSize: 13, fontWeight: 600 }}>
              Activity
              {runningCount > 0 && (
                <span
                  style={{
                    marginLeft: 8,
                    background: "#e8a317",
                    color: "#000",
                    borderRadius: 10,
                    padding: "1px 8px",
                    fontSize: 11,
                  }}
                >
                  {runningCount}
                </span>
              )}
            </span>
            {onToggle && (
              <button
                onClick={onToggle}
                style={{
                  background: "none",
                  border: "none",
                  color: "#a0a0b0",
                  cursor: "pointer",
                  fontSize: 16,
                }}
              >
                ✕
              </button>
            )}
          </div>

          {/* Tool cards */}
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            {toolCards.map((card, i) => (
              <div
                key={i}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 8,
                  padding: "8px 10px",
                  background: "#1a1a2e",
                  borderRadius: 8,
                  fontSize: 13,
                }}
              >
                {/* Status indicator */}
                <span style={{ fontSize: 14 }}>
                  {card.status === "running"
                    ? "⏳"
                    : card.status === "done"
                      ? "✓"
                      : "✗"}
                </span>

                {/* Name */}
                <span
                  style={{
                    flex: 1,
                    color: "#e0e0e0",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  }}
                >
                  {card.displayName}
                </span>

                {/* Elapsed time */}
                {card.elapsed && (
                  <span style={{ color: "#666", fontSize: 11 }}>
                    {card.elapsed}
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
