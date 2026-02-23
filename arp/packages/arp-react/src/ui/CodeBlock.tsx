import { useState, useCallback } from "react";
import type { CodeBlockProps } from "@haira/arp";

export function CodeBlock({
  title,
  language,
  code,
  tabs,
}: CodeBlockProps & { onInput?: (text: string) => void }) {
  const [activeTab, setActiveTab] = useState(0);
  const [copied, setCopied] = useState(false);

  const activeCode = tabs?.[activeTab]?.code ?? code ?? "";
  const activeLang = tabs?.[activeTab]?.language ?? language ?? "";

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(activeCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }, [activeCode]);

  return (
    <div className="arp-code-block" style={{ background: "#111118", borderRadius: 8, overflow: "hidden" }}>
      <div style={{ display: "flex", alignItems: "center", padding: "8px 12px", borderBottom: "1px solid #1a1a2e", gap: 8 }}>
        {title && <span style={{ fontSize: 13, fontWeight: 600, color: "#e0e0e0" }}>{title}</span>}
        {activeLang && <span style={{ fontSize: 11, color: "#666", marginLeft: title ? 0 : "auto" }}>{activeLang}</span>}
        {tabs && tabs.length > 1 && (
          <div style={{ display: "flex", gap: 4, marginLeft: 8 }}>
            {tabs.map((tab, i) => (
              <button
                key={i}
                onClick={() => setActiveTab(i)}
                style={{
                  background: i === activeTab ? "#2a2a3e" : "transparent",
                  color: i === activeTab ? "#e0e0e0" : "#666",
                  border: "none",
                  borderRadius: 6,
                  padding: "3px 8px",
                  fontSize: 11,
                  cursor: "pointer",
                }}
              >
                {tab.name}
              </button>
            ))}
          </div>
        )}
        <button
          onClick={handleCopy}
          style={{
            marginLeft: "auto",
            background: "none",
            border: "none",
            color: copied ? "#22c55e" : "#666",
            cursor: "pointer",
            fontSize: 12,
          }}
        >
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
      <pre style={{ margin: 0, padding: "12px 16px", overflow: "auto", maxHeight: 480, fontSize: 13, lineHeight: 1.5, color: "#e0e0e0" }}>
        <code>{activeCode}</code>
      </pre>
    </div>
  );
}
