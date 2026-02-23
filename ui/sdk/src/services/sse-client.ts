// SSE streaming client — re-exports from @haira/arp + Haira-specific submitForm.

export { streamSSE, connectSSE } from "@haira/arp";
export type { SSECallbacks, ToolEvent } from "@haira/arp";

/** Non-streaming form submission */
export async function submitForm(
  url: string,
  method: string,
  body: unknown,
  hasFile: boolean,
  formData?: FormData
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
