import type { StepEvent, ToolRenderEvent } from "./types";

export interface ToolEvent {
  tool: string;
  args?: string;
  ok?: boolean;
}

export interface SSECallbacks {
  onStep?: (event: StepEvent) => void;
  onDelta?: (delta: string) => void;
  onResult?: (data: unknown) => void;
  onError?: (error: string) => void;
  onDone?: () => void;
  onToolStart?: (event: ToolEvent) => void;
  onToolEnd?: (event: ToolEvent) => void;
  onToolRender?: (event: ToolRenderEvent) => void;
}

export async function streamSSE(
  url: string,
  body: unknown,
  callbacks: SSECallbacks,
  formData?: FormData,
): Promise<void> {
  const headers: Record<string, string> = {
    Accept: "text/event-stream",
  };
  let reqBody: BodyInit;
  if (formData) {
    reqBody = formData;
  } else {
    headers["Content-Type"] = "application/json";
    reqBody = JSON.stringify(body);
  }
  const resp = await fetch(url, {
    method: "POST",
    headers,
    body: reqBody,
  });

  if (!resp.ok) {
    const text = await resp.text();
    callbacks.onError?.(text || `HTTP ${resp.status}`);
    callbacks.onDone?.();
    return;
  }

  const reader = resp.body!.getReader();
  const decoder = new TextDecoder();
  let buf = "";
  let currentEvent = "";

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
            // Bare data: lines (legacy protocol) — treat as delta
            if (parsed.delta) {
              callbacks.onDelta?.(parsed.delta);
            }
            break;
        }
      } catch {
        // Non-JSON data line, ignore
      }

      currentEvent = "";
    }
  }

  callbacks.onDone?.();
}

export async function submitForm(
  url: string,
  method: string,
  body: unknown,
  hasFile: boolean,
  formData?: FormData,
): Promise<{ status: number; data: unknown }> {
  let opts: RequestInit;

  if (method === "GET" || method === "DELETE") {
    const qs = new URLSearchParams(body as Record<string, string>);
    const fullUrl = qs.toString() ? `${url}?${qs}` : url;
    opts = { method };
    url = fullUrl;
  } else if (hasFile && formData) {
    opts = { method, body: formData };
  } else {
    opts = {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    };
  }

  const resp = await fetch(url, opts);
  const text = await resp.text();
  let data: unknown;
  try {
    data = JSON.parse(text);
  } catch {
    data = text;
  }
  return { status: resp.status, data };
}
