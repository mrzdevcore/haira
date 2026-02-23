import type { ToolRenderEvent } from "@haira/arp";
import { useArpContext } from "../context/ArpProvider.js";

// ---------------------------------------------------------------------------
// Props
// ---------------------------------------------------------------------------

export interface ArpRendererProps {
  /** The render event to display. */
  event: ToolRenderEvent;
  /** Custom component registry (merged with context components). */
  components?: Record<string, React.ComponentType<any>>;
  /** Max nesting depth for group components (default: 3). */
  maxDepth?: number;
  /** Whether this component was restored from a session. */
  restored?: boolean;
  /** Callback for user interactions (confirm, choices, form submit). */
  onInput?: (text: string) => void;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

/**
 * Renders an ARP component by mapping the component type name to a
 * registered React component. Falls back to a JSON dump for unknown types.
 */
export function ArpRenderer({
  event,
  components: localComponents,
  maxDepth = 3,
  restored = false,
  onInput,
}: ArpRendererProps) {
  const { components: contextComponents } = useArpContext();

  // Merge: local overrides > context > empty
  const registry = { ...contextComponents, ...localComponents };

  const Component = registry[event.component];

  if (!Component) {
    // Unknown component — render props as JSON
    return (
      <pre
        style={{
          background: "#1a1a2e",
          color: "#a0a0b0",
          padding: "12px",
          borderRadius: "8px",
          fontSize: "12px",
          overflow: "auto",
        }}
      >
        {JSON.stringify({ component: event.component, props: event.props }, null, 2)}
      </pre>
    );
  }

  return (
    <Component
      {...event.props}
      _restored={restored}
      onInput={onInput}
    />
  );
}
