import type { ToolRenderEvent } from "@haira/arp";
import { ArpRenderer } from "./ArpRenderer.js";

// ---------------------------------------------------------------------------
// Props
// ---------------------------------------------------------------------------

export interface ArpMessageProps {
  role: "user" | "assistant";
  content: string;
  /** Optional avatar URL or emoji for assistant messages. */
  avatar?: string;
  /** Filename badge for file attachments. */
  file?: string;
  /** Generative UI components rendered after text. */
  uiEvents?: ToolRenderEvent[];
  /** Whether this message was restored from session history. */
  restored?: boolean;
  /** Custom markdown renderer. If not provided, uses simple HTML rendering. */
  renderMarkdown?: (content: string) => React.ReactNode;
  /** Callback for user interactions from UI components. */
  onInput?: (text: string) => void;
}

// ---------------------------------------------------------------------------
// Simple markdown-to-HTML converter
// ---------------------------------------------------------------------------

function defaultRenderMarkdown(content: string): React.ReactNode {
  if (!content) return null;

  // Very basic: convert code blocks, bold, italic, links, and newlines.
  // For production, users should provide a proper markdown renderer.
  const html = content
    // Code blocks
    .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre style="background:#1a1a2e;color:#e0e0e0;padding:12px;border-radius:8px;overflow:auto;font-size:13px"><code>$2</code></pre>')
    // Inline code
    .replace(/`([^`]+)`/g, '<code style="background:#2a2a3e;padding:2px 6px;border-radius:4px;font-size:13px">$1</code>')
    // Bold
    .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
    // Italic
    .replace(/\*(.+?)\*/g, "<em>$1</em>")
    // Links
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>')
    // Line breaks (double newline = paragraph)
    .replace(/\n\n/g, "</p><p>")
    .replace(/\n/g, "<br/>");

  return (
    <div
      className="arp-message-content"
      dangerouslySetInnerHTML={{ __html: `<p>${html}</p>` }}
    />
  );
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

/**
 * Renders a single chat message (user or assistant) with optional
 * markdown content, file badges, and generative UI components.
 */
export function ArpMessage({
  role,
  content,
  avatar,
  file,
  uiEvents,
  restored = false,
  renderMarkdown = defaultRenderMarkdown,
  onInput,
}: ArpMessageProps) {
  const isUser = role === "user";

  return (
    <div
      className={`arp-message arp-message--${role}`}
      style={{
        display: "flex",
        gap: "12px",
        padding: "16px 0",
        flexDirection: isUser ? "row-reverse" : "row",
      }}
    >
      {/* Avatar */}
      {!isUser && (
        <div
          className="arp-message-avatar"
          style={{
            width: 32,
            height: 32,
            borderRadius: "50%",
            background: "#2a2a3e",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            flexShrink: 0,
            fontSize: 14,
            color: "#a0a0b0",
          }}
        >
          {avatar || "A"}
        </div>
      )}

      {/* Content */}
      <div
        style={{
          maxWidth: "80%",
          minWidth: 0,
        }}
      >
        {/* File badge */}
        {file && (
          <div
            className="arp-message-file"
            style={{
              fontSize: 12,
              color: "#e8a317",
              marginBottom: 4,
            }}
          >
            📎 {file}
          </div>
        )}

        {/* Text content */}
        {content && (
          <div
            className="arp-message-bubble"
            style={{
              background: isUser ? "#2a2a3e" : "transparent",
              padding: isUser ? "10px 14px" : "0",
              borderRadius: isUser ? "12px" : "0",
              color: "#e0e0e0",
              fontSize: 14,
              lineHeight: 1.6,
            }}
          >
            {renderMarkdown(content)}
          </div>
        )}

        {/* Generative UI components */}
        {uiEvents?.map((event, i) => (
          <div key={i} style={{ marginTop: 8 }}>
            <ArpRenderer
              event={event}
              restored={restored}
              onInput={onInput}
            />
          </div>
        ))}
      </div>
    </div>
  );
}
