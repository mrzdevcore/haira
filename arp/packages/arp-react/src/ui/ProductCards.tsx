import { useState, useMemo } from "react";
import type { ProductCardsProps } from "@haira/arp";

export function ProductCards({
  title,
  cards,
}: ProductCardsProps & { onInput?: (text: string) => void }) {
  const [search, setSearch] = useState("");
  const showSearch = cards.length >= 12;

  const filtered = useMemo(() => {
    if (!search) return cards;
    const q = search.toLowerCase();
    return cards.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        c.brand?.toLowerCase().includes(q) ||
        c.description?.toLowerCase().includes(q) ||
        c.price.toLowerCase().includes(q),
    );
  }, [cards, search]);

  return (
    <div className="arp-product-cards" style={{ background: "#111118", borderRadius: 8, overflow: "hidden" }}>
      <div style={{ padding: "12px 16px", display: "flex", alignItems: "center", gap: 12, borderBottom: "1px solid #1a1a2e" }}>
        {title && <span style={{ fontWeight: 600, fontSize: 14, color: "#e0e0e0" }}>{title}</span>}
        {showSearch && (
          <input
            type="text"
            placeholder="Search..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            style={{
              marginLeft: "auto",
              background: "#1a1a2e",
              color: "#e0e0e0",
              border: "1px solid #2a2a3e",
              borderRadius: 6,
              padding: "4px 10px",
              fontSize: 12,
              width: 160,
            }}
          />
        )}
      </div>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fill, minmax(190px, 1fr))",
          gap: 12,
          padding: 16,
          maxHeight: 520,
          overflow: "auto",
        }}
      >
        {filtered.map((card, i) => (
          <div
            key={i}
            style={{
              background: "#1a1a2e",
              borderRadius: 8,
              overflow: "hidden",
              cursor: card.url ? "pointer" : "default",
            }}
            onClick={() => card.url && window.open(card.url, "_blank")}
          >
            {card.image && (
              <div style={{ height: 140, overflow: "hidden", background: "#0a0a12" }}>
                <img
                  src={card.image}
                  alt={card.name}
                  loading="lazy"
                  style={{ width: "100%", height: "100%", objectFit: "cover" }}
                />
              </div>
            )}
            <div style={{ padding: "10px 12px" }}>
              {card.badge && (
                <span
                  style={{
                    fontSize: 10,
                    background: "#e8a317",
                    color: "#000",
                    borderRadius: 4,
                    padding: "1px 6px",
                    fontWeight: 600,
                    marginBottom: 4,
                    display: "inline-block",
                  }}
                >
                  {card.badge}
                </span>
              )}
              <div style={{ fontWeight: 600, fontSize: 13, color: "#e0e0e0", marginTop: 2 }}>
                {card.name}
              </div>
              {card.brand && (
                <div style={{ fontSize: 11, color: "#666" }}>{card.brand}</div>
              )}
              {card.description && (
                <div style={{ fontSize: 12, color: "#888", marginTop: 4, lineHeight: 1.4 }}>
                  {card.description}
                </div>
              )}
              <div style={{ fontWeight: 700, fontSize: 14, color: "#e8a317", marginTop: 6 }}>
                {card.price}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
