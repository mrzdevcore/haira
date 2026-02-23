import {
  useState,
  useRef,
  useEffect,
  useCallback,
  type CSSProperties,
  type KeyboardEvent,
} from "react";
import { useArpChat, type UseArpChatOptions } from "../hooks/use-arp-chat.js";
import { ArpProvider } from "../context/ArpProvider.js";
import { ArpMessage } from "./ArpMessage.js";
import { ArpActivityPanel } from "./ArpActivityPanel.js";

// ---------------------------------------------------------------------------
// Props
// ---------------------------------------------------------------------------

export interface ArpChatProps extends UseArpChatOptions {
  /** Workflow title displayed in the header. */
  title?: string;
  /** Workflow description for the welcome screen. */
  description?: string;
  /** Suggested prompts for the welcome screen. */
  suggestions?: string[];
  /** Assistant avatar (URL or emoji). */
  avatar?: string;
  /** Custom logo URL for the header. */
  logo?: string;
  /** Theme: "dark" (default) or "light". */
  theme?: "dark" | "light";
  /** Accent color (default: "#e8a317"). */
  accentColor?: string;
  /** Custom component registry (extends or replaces built-ins). */
  components?: Record<string, React.ComponentType<any>>;
  /** CSS class for the root element. */
  className?: string;
  /** Inline styles for the root element. */
  style?: CSSProperties;
  /** Show activity panel (default: true). */
  showActivityPanel?: boolean;
  /** Custom markdown renderer for messages. */
  renderMarkdown?: (content: string) => React.ReactNode;
  /** Callback when a message is sent. */
  onSend?: (message: string) => void;
  /** Callback when session changes. */
  onSessionChange?: (sessionId: string) => void;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

/**
 * Full-featured chat UI component for the Agent Rendering Protocol.
 *
 * Drop-in usage:
 * ```tsx
 * <ArpChat
 *   url="ws://localhost:8080/_arp/v1"
 *   title="My Assistant"
 *   suggestions={["What can you do?"]}
 * />
 * ```
 */
export function ArpChat({
  title = "Chat",
  description,
  suggestions,
  avatar,
  logo,
  theme = "dark",
  accentColor = "#e8a317",
  components = {},
  className,
  style,
  showActivityPanel = true,
  renderMarkdown,
  onSend,
  onSessionChange,
  ...chatOptions
}: ArpChatProps) {
  const chat = useArpChat(chatOptions);
  const [input, setInput] = useState("");
  const [panelOpen, setPanelOpen] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const isDark = theme === "dark";

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [chat.messages]);

  // Open activity panel when tools start
  useEffect(() => {
    if (chat.runningToolCount > 0 && showActivityPanel) {
      setPanelOpen(true);
    }
  }, [chat.runningToolCount, showActivityPanel]);

  // Close activity panel when done
  useEffect(() => {
    if (!chat.isStreaming && panelOpen) {
      const timer = setTimeout(() => setPanelOpen(false), 500);
      return () => clearTimeout(timer);
    }
  }, [chat.isStreaming, panelOpen]);

  // Notify parent of session changes
  useEffect(() => {
    onSessionChange?.(chat.sessionId);
  }, [chat.sessionId, onSessionChange]);

  const handleSend = useCallback(() => {
    const text = input.trim();
    if (!text || chat.isStreaming) return;
    setInput("");
    onSend?.(text);
    chat.sendMessage(text);
    // Reset textarea height
    if (inputRef.current) inputRef.current.style.height = "auto";
  }, [input, chat, onSend]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend],
  );

  const handleSuggestion = useCallback(
    (text: string) => {
      setInput("");
      onSend?.(text);
      chat.sendMessage(text);
    },
    [chat, onSend],
  );

  const handleInput = useCallback(
    (text: string) => {
      chat.sendMessage(text);
    },
    [chat],
  );

  // Auto-resize textarea
  const handleTextareaChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      setInput(e.target.value);
      e.target.style.height = "auto";
      e.target.style.height = `${Math.min(e.target.scrollHeight, 160)}px`;
    },
    [],
  );

  const showWelcome = chat.messages.length === 0;

  // --- Styles ---
  const rootStyle: CSSProperties = {
    display: "flex",
    height: "100%",
    background: isDark ? "#09090b" : "#ffffff",
    color: isDark ? "#e0e0e0" : "#1a1a1a",
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    ...style,
  };

  return (
    <ArpProvider
      connected={chat.isConnected}
      capabilities={chat.capabilities}
      components={components}
    >
      <div className={`arp-chat ${className ?? ""}`} style={rootStyle}>
        {/* Main chat area */}
        <div style={{ flex: 1, display: "flex", flexDirection: "column", minWidth: 0 }}>
          {/* Header */}
          <div
            className="arp-chat-header"
            style={{
              padding: "12px 20px",
              borderBottom: `1px solid ${isDark ? "#1a1a2e" : "#e5e5e5"}`,
              display: "flex",
              alignItems: "center",
              gap: 10,
            }}
          >
            {logo && (
              <img src={logo} alt="" style={{ height: 24, width: 24 }} />
            )}
            <span style={{ fontWeight: 600, fontSize: 15 }}>{title}</span>
            {chat.isConnected && (
              <span
                style={{
                  width: 8,
                  height: 8,
                  borderRadius: "50%",
                  background: "#22c55e",
                  marginLeft: 4,
                }}
              />
            )}
          </div>

          {/* Messages area */}
          <div
            className="arp-chat-messages"
            style={{
              flex: 1,
              overflow: "auto",
              padding: "0 20px",
            }}
          >
            {/* Welcome screen */}
            {showWelcome && (
              <div
                style={{
                  display: "flex",
                  flexDirection: "column",
                  alignItems: "center",
                  justifyContent: "center",
                  height: "100%",
                  gap: 16,
                  textAlign: "center",
                }}
              >
                <div style={{ fontSize: 28, fontWeight: 700, color: accentColor }}>
                  {title}
                </div>
                {description && (
                  <div style={{ color: isDark ? "#888" : "#666", maxWidth: 480 }}>
                    {description}
                  </div>
                )}
                {suggestions && suggestions.length > 0 && (
                  <div
                    style={{
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 8,
                      justifyContent: "center",
                      marginTop: 8,
                    }}
                  >
                    {suggestions.map((s, i) => (
                      <button
                        key={i}
                        onClick={() => handleSuggestion(s)}
                        style={{
                          background: isDark ? "#1a1a2e" : "#f0f0f0",
                          color: isDark ? "#ccc" : "#333",
                          border: `1px solid ${isDark ? "#2a2a3e" : "#ddd"}`,
                          borderRadius: 20,
                          padding: "8px 16px",
                          cursor: "pointer",
                          fontSize: 13,
                          transition: "background 0.15s",
                        }}
                      >
                        {s}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Message list */}
            {chat.messages.map((msg) => (
              <ArpMessage
                key={msg.id}
                role={msg.role}
                content={msg.content}
                avatar={avatar}
                file={msg.file}
                uiEvents={msg.uiEvents}
                restored={msg.restored}
                renderMarkdown={renderMarkdown}
                onInput={handleInput}
              />
            ))}

            {/* Typing indicator */}
            {chat.isStreaming && chat.messages.length > 0 && !chat.messages[chat.messages.length - 1]?.content && (
              <div style={{ padding: "8px 0", color: "#666", fontSize: 13 }}>
                Thinking...
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>

          {/* Input area */}
          <div
            className="arp-chat-input"
            style={{
              padding: "12px 20px",
              borderTop: `1px solid ${isDark ? "#1a1a2e" : "#e5e5e5"}`,
            }}
          >
            <div
              style={{
                display: "flex",
                gap: 10,
                alignItems: "flex-end",
                background: isDark ? "#111118" : "#f8f8f8",
                borderRadius: 12,
                padding: "8px 12px",
                border: `1px solid ${isDark ? "#2a2a3e" : "#ddd"}`,
              }}
            >
              <textarea
                ref={inputRef}
                value={input}
                onChange={handleTextareaChange}
                onKeyDown={handleKeyDown}
                placeholder="Type a message..."
                disabled={chat.isStreaming}
                rows={1}
                style={{
                  flex: 1,
                  background: "transparent",
                  border: "none",
                  outline: "none",
                  color: isDark ? "#e0e0e0" : "#1a1a1a",
                  fontSize: 14,
                  lineHeight: 1.5,
                  resize: "none",
                  fontFamily: "inherit",
                  maxHeight: 160,
                }}
              />
              <button
                onClick={handleSend}
                disabled={chat.isStreaming || !input.trim()}
                style={{
                  background: accentColor,
                  color: "#000",
                  border: "none",
                  borderRadius: 8,
                  padding: "6px 14px",
                  cursor: chat.isStreaming || !input.trim() ? "default" : "pointer",
                  opacity: chat.isStreaming || !input.trim() ? 0.4 : 1,
                  fontWeight: 600,
                  fontSize: 13,
                  flexShrink: 0,
                  transition: "opacity 0.15s",
                }}
              >
                Send
              </button>
            </div>
          </div>
        </div>

        {/* Activity panel */}
        {showActivityPanel && (
          <ArpActivityPanel
            toolCards={chat.toolCards}
            open={panelOpen}
            onToggle={() => setPanelOpen(!panelOpen)}
          />
        )}
      </div>
    </ArpProvider>
  );
}
