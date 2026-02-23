/**
 * ARP Component Registry — framework-agnostic interface for mapping
 * ARP component type names to framework-specific implementations.
 *
 * Each framework package (@haira/arp-react, @haira/arp-vue, etc.)
 * provides its own registry implementation with built-in components.
 */

// ---------------------------------------------------------------------------
// Registry interface
// ---------------------------------------------------------------------------

/**
 * A registry that maps ARP component type names (e.g., "status-card", "table")
 * to framework-specific component implementations.
 *
 * @typeParam T - The component type (e.g., React.ComponentType, Vue Component)
 */
export interface ComponentRegistry<T = unknown> {
  /** Register a component for a given ARP component type name. */
  register(typeName: string, component: T): void;
  /** Get the component for a given ARP component type name. */
  get(typeName: string): T | undefined;
  /** List all registered component type names. */
  list(): string[];
  /** Check if a component type is registered. */
  has(typeName: string): boolean;
}

/**
 * Create a simple in-memory component registry.
 *
 * @param initial - Optional initial mappings.
 */
export function createComponentRegistry<T>(
  initial?: Record<string, T>,
): ComponentRegistry<T> {
  const map = new Map<string, T>(
    initial ? Object.entries(initial) : undefined,
  );

  return {
    register(typeName: string, component: T): void {
      map.set(typeName, component);
    },
    get(typeName: string): T | undefined {
      return map.get(typeName);
    },
    list(): string[] {
      return Array.from(map.keys());
    },
    has(typeName: string): boolean {
      return map.has(typeName);
    },
  };
}

// ---------------------------------------------------------------------------
// Built-in component names
// ---------------------------------------------------------------------------

/** Standard ARP component type names supported by the protocol. */
export const BUILT_IN_COMPONENTS = [
  "text",
  "status-card",
  "table",
  "code-block",
  "diff",
  "key-value",
  "progress",
  "chart",
  "form",
  "confirm",
  "choices",
  "product-cards",
  "markdown",
  "image",
] as const;

export type BuiltInComponentType = (typeof BUILT_IN_COMPONENTS)[number];
