import { useState, useCallback } from "react";
import type { FormViewProps } from "@haira/arp";

export function Form({
  title,
  fields,
  submit_label = "Submit",
  submit_action,
  onInput,
}: FormViewProps & { onInput?: (text: string) => void; _restored?: boolean }) {
  const [values, setValues] = useState<Record<string, string>>(() => {
    const init: Record<string, string> = {};
    for (const f of fields) {
      init[f.name] = f.value ?? "";
    }
    return init;
  });

  const handleChange = useCallback((name: string, value: string) => {
    setValues((prev) => ({ ...prev, [name]: value }));
  }, []);

  const handleSubmit = useCallback(() => {
    const payload = submit_action
      ? `[Form: ${submit_action}] ${JSON.stringify(values)}`
      : JSON.stringify(values);
    onInput?.(payload);
  }, [values, submit_action, onInput]);

  return (
    <div className="arp-form" style={{ background: "#111118", borderRadius: 8, padding: "16px" }}>
      {title && (
        <div style={{ fontWeight: 600, fontSize: 14, color: "#e0e0e0", marginBottom: 12 }}>
          {title}
        </div>
      )}
      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {fields.map((field) => (
          <div key={field.name}>
            <label style={{ display: "block", fontSize: 12, color: "#888", marginBottom: 4 }}>
              {field.label ?? field.name}
              {field.required && <span style={{ color: "#ef4444" }}> *</span>}
            </label>
            {field.options ? (
              <select
                value={values[field.name] ?? ""}
                onChange={(e) => handleChange(field.name, e.target.value)}
                style={{
                  width: "100%",
                  padding: "8px 10px",
                  background: "#1a1a2e",
                  color: "#e0e0e0",
                  border: "1px solid #2a2a3e",
                  borderRadius: 6,
                  fontSize: 13,
                }}
              >
                <option value="">Select...</option>
                {field.options.map((opt) => (
                  <option key={opt} value={opt}>{opt}</option>
                ))}
              </select>
            ) : (
              <input
                type={field.field_type === "number" ? "number" : "text"}
                value={values[field.name] ?? ""}
                onChange={(e) => handleChange(field.name, e.target.value)}
                style={{
                  width: "100%",
                  padding: "8px 10px",
                  background: "#1a1a2e",
                  color: "#e0e0e0",
                  border: "1px solid #2a2a3e",
                  borderRadius: 6,
                  fontSize: 13,
                  boxSizing: "border-box",
                }}
              />
            )}
          </div>
        ))}
      </div>
      <button
        onClick={handleSubmit}
        style={{
          marginTop: 16,
          background: "#e8a317",
          color: "#000",
          border: "none",
          borderRadius: 8,
          padding: "8px 20px",
          fontWeight: 600,
          fontSize: 13,
          cursor: "pointer",
        }}
      >
        {submit_label}
      </button>
    </div>
  );
}
