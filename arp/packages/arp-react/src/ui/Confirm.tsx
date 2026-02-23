import { useState, useCallback } from "react";
import type { ConfirmProps } from "@haira/arp";

export function Confirm({
  title,
  message,
  confirm_label = "Confirm",
  deny_label = "Deny",
  _restored,
  onInput,
}: ConfirmProps & { onInput?: (text: string) => void }) {
  const [chosen, setChosen] = useState<string | null>(null);

  const handleChoice = useCallback(
    (label: string, action: string) => {
      setChosen(label);
      onInput?.(`[${action}] ${label}`);
    },
    [onInput],
  );

  const isDisabled = !!chosen || !!_restored;

  return (
    <div className="arp-confirm" style={{ background: "#111118", borderRadius: 8, padding: "16px" }}>
      <div style={{ fontWeight: 600, fontSize: 14, color: "#e0e0e0" }}>{title}</div>
      {message && (
        <div style={{ marginTop: 6, fontSize: 13, color: "#a0a0b0", lineHeight: 1.5 }}>
          {message}
        </div>
      )}
      {isDisabled ? (
        <div style={{ marginTop: 12, fontSize: 12, color: "#666" }}>
          Selection made
        </div>
      ) : (
        <div style={{ marginTop: 12, display: "flex", gap: 8 }}>
          <button
            onClick={() => handleChoice(deny_label, "Denied")}
            style={{
              background: "#1a1a2e",
              color: "#a0a0b0",
              border: "1px solid #2a2a3e",
              borderRadius: 8,
              padding: "8px 16px",
              cursor: "pointer",
              fontSize: 13,
            }}
          >
            {deny_label}
          </button>
          <button
            onClick={() => handleChoice(confirm_label, "Confirmed")}
            style={{
              background: "#e8a317",
              color: "#000",
              border: "none",
              borderRadius: 8,
              padding: "8px 16px",
              cursor: "pointer",
              fontWeight: 600,
              fontSize: 13,
            }}
          >
            {confirm_label}
          </button>
        </div>
      )}
    </div>
  );
}
