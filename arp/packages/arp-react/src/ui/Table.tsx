import { useState } from "react";
import type { TableProps } from "@haira/arp";

export function Table({
  title,
  headers,
  rows,
  tabs,
  highlight,
}: TableProps & { onInput?: (text: string) => void }) {
  const [activeTab, setActiveTab] = useState(0);

  const activeHeaders = tabs?.[activeTab]?.headers ?? headers;
  const activeRows = tabs?.[activeTab]?.rows ?? rows;
  const activeHighlight = tabs?.[activeTab]?.highlight ?? highlight;

  return (
    <div className="arp-table" style={{ background: "#111118", borderRadius: 8, overflow: "hidden" }}>
      {(title || tabs) && (
        <div style={{ padding: "10px 16px", borderBottom: "1px solid #1a1a2e", display: "flex", alignItems: "center", gap: 12 }}>
          {title && <span style={{ fontWeight: 600, fontSize: 14, color: "#e0e0e0" }}>{title}</span>}
          {tabs && tabs.length > 1 && (
            <div style={{ display: "flex", gap: 4, marginLeft: "auto" }}>
              {tabs.map((tab, i) => (
                <button
                  key={i}
                  onClick={() => setActiveTab(i)}
                  style={{
                    background: i === activeTab ? "#2a2a3e" : "transparent",
                    color: i === activeTab ? "#e0e0e0" : "#666",
                    border: "none",
                    borderRadius: 6,
                    padding: "4px 10px",
                    fontSize: 12,
                    cursor: "pointer",
                  }}
                >
                  {tab.name}
                </button>
              ))}
            </div>
          )}
        </div>
      )}
      <div style={{ overflowX: "auto" }}>
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
          <thead>
            <tr>
              {activeHeaders.map((h, i) => (
                <th key={i} style={{ padding: "8px 12px", textAlign: "left", color: "#888", fontWeight: 500, borderBottom: "1px solid #1a1a2e", whiteSpace: "nowrap" }}>
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {activeRows.map((row, ri) => (
              <tr
                key={ri}
                style={{
                  background: activeHighlight?.includes(ri) ? "rgba(232,163,23,0.08)" : undefined,
                }}
              >
                {row.map((cell, ci) => (
                  <td key={ci} style={{ padding: "8px 12px", color: "#e0e0e0", borderBottom: "1px solid #0a0a12" }}>
                    {cell}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
