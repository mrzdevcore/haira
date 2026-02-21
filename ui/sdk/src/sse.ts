// Backward compatibility — re-exports from services/sse-client.
// Existing components import from here; new code should import from "services".

export { streamSSE, connectSSE, submitForm } from "./services/sse-client";
export type { SSECallbacks, ToolEvent } from "./services/sse-client";
