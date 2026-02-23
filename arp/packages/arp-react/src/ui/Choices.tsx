import { useState, useCallback } from "react";
import type { ChoicesProps } from "@haira/arp";

export function Choices({
  title,
  options,
  style: variant = "buttons",
  _restored,
  onInput,
}: ChoicesProps & { onInput?: (text: string) => void }) {
  const [selected, setSelected] = useState<string | null>(null);

  const handleSelect = useCallback(
    (option: string) => {
      setSelected(option);
      onInput?.(option);
    },
    [onInput],
  );

  const isDisabled = !!selected || !!_restored;

  return (
    <div className="arp-choices" style={{ background: "#111118", borderRadius: 8, padding: "16px" }}>
      <div style={{ fontWeight: 600, fontSize: 14, color: "#e0e0e0", marginBottom: 10 }}>
        {title}
      </div>
      {isDisabled ? (
        <div style={{ fontSize: 12, color: "#666" }}>
          Selected: {selected ?? "(from history)"}
        </div>
      ) : variant === "list" ? (
        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          {options.map((opt, i) => (
            <button
              key={i}
              onClick={() => handleSelect(opt)}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 10,
                background: "#1a1a2e",
                color: "#e0e0e0",
                border: "1px solid #2a2a3e",
                borderRadius: 8,
                padding: "10px 14px",
                cursor: "pointer",
                fontSize: 13,
                textAlign: "left",
              }}
            >
              <span style={{ width: 16, height: 16, borderRadius: "50%", border: "2px solid #444", flexShrink: 0 }} />
              {opt}
            </button>
          ))}
        </div>
      ) : (
        <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
          {options.map((opt, i) => (
            <button
              key={i}
              onClick={() => handleSelect(opt)}
              style={{
                background: "#1a1a2e",
                color: "#e0e0e0",
                border: "1px solid #2a2a3e",
                borderRadius: 20,
                padding: "8px 16px",
                cursor: "pointer",
                fontSize: 13,
              }}
            >
              {opt}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
