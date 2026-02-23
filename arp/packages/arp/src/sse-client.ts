/**
 * ARP SSE Client — Server-Sent Events transport for the Agent Rendering Protocol.
 *
 * Provides POST-based SSE streaming (for sending messages) and GET-based SSE
 * reconnection (for resuming in-progress streams). Uses fetch API only — no
 * EventSource (which doesn't support POST).
 */

import type { StepEvent, ToolRenderEvent } from "./types.js";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface ToolEvent {
  tool: string;
  args?: string;
  ok?: boolean;
}

export interface SSECallbacks {
  onRunId?: (id: string) => void;
  onStep?: (event: StepEvent) => void;
  onDelta?: (delta: string) => void;
  onResult?: (data: unknown) => void;
  onError?: (error: string) => void;
  onDone?: () => void;
  onToolStart?: (event: ToolEvent) => void;
  onToolEnd?: (event: ToolEvent) => void;
  onToolRender?: (event: ToolRenderEvent) => void;
}

export interface StreamSSEOptions {
  /** FormData for multipart uploads (overrides JSON body). */
  formData?: FormData;
  /** AbortSignal for cancellation. */
  signal?: AbortSignal;
  /** Additional request headers. */
  headers?: Record<string, string>;
}

// ---------------------------------------------------------------------------
// SSE Parser
// ---------------------------------------------------------------------------

async function parseSSEStream(
  reader: ReadableStreamDefaultReader<Uint8Array>,
  callbacks: SSECallbacks,
  signal?: AbortSignal,
): Promise<void> {
  const decoder = new TextDecoder();
  let buf = "";
  let currentEvent = "";

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buf += decoder.decode(value, { stream: true });
      const lines = buf.split("\n");
      buf = lines.pop()!;

      for (const line of lines) {
        const trimmed = line.trim();

        if (trimmed.startsWith("event:")) {
          currentEvent = trimmed.slice(6).trim();
          continue;
        }

        if (!trimmed.startsWith("data:")) continue;

        const data = trimmed.slice(5).trim();
        if (data === "[DONE]") {
          callbacks.onDone?.();
          return;
        }

        try {
          const parsed = JSON.parse(data);

          switch (currentEvent) {
            case "run_id":
              callbacks.onRunId?.(parsed.run_id);
              break;
            case "step":
              callbacks.onStep?.(parsed as StepEvent);
              break;
            case "result":
              callbacks.onResult?.(parsed);
              break;
            case "error":
              callbacks.onError?.(parsed.error || "Unknown error");
              break;
            case "delta":
              callbacks.onDelta?.(parsed.delta);
              break;
            case "tool_start":
              callbacks.onToolStart?.(parsed as ToolEvent);
              break;
            case "tool_end":
              callbacks.onToolEnd?.(parsed as ToolEvent);
              break;
            case "tool_render":
              callbacks.onToolRender?.(parsed as ToolRenderEvent);
              break;
            default:
              if (parsed.delta) {
                callbacks.onDelta?.(parsed.delta);
              }
              break;
          }
        } catch {
          // Non-JSON data line — skip
        }

        currentEvent = "";
      }
    }

    callbacks.onDone?.();
  } catch (err) {
    if (signal?.aborted) return;
    throw err;
  }
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/** POST-based SSE stream for sending messages/data to the agent. */
export async function streamSSE(
  url: string,
  body: unknown,
  callbacks: SSECallbacks,
  options?: StreamSSEOptions,
): Promise<void> {
  const headers: Record<string, string> = {
    Accept: "text/event-stream",
    ...options?.headers,
  };

  let reqBody: BodyInit;
  if (options?.formData) {
    reqBody = options.formData;
  } else {
    headers["Content-Type"] = "application/json";
    reqBody = JSON.stringify(body);
  }

  const resp = await fetch(url, {
    method: "POST",
    headers,
    body: reqBody,
    signal: options?.signal,
  });

  if (!resp.ok) {
    const text = await resp.text();
    callbacks.onError?.(text || `HTTP ${resp.status}`);
    callbacks.onDone?.();
    return;
  }

  await parseSSEStream(resp.body!.getReader(), callbacks, options?.signal);
}

/** GET-based SSE for reconnecting to an in-progress stream. */
export async function connectSSE(
  url: string,
  callbacks: SSECallbacks,
  signal?: AbortSignal,
): Promise<void> {
  const resp = await fetch(url, {
    method: "GET",
    headers: { Accept: "text/event-stream" },
    signal,
  });

  if (!resp.ok) {
    const text = await resp.text();
    callbacks.onError?.(text || `HTTP ${resp.status}`);
    callbacks.onDone?.();
    return;
  }

  await parseSSEStream(resp.body!.getReader(), callbacks, signal);
}
