const API_BASE = '/api'

export interface ApiMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
}

export async function createConversation(userId: string, message: string): Promise<{ conversationId: string; reply: ChatMessage }> {
  const res = await fetch(`${API_BASE}/conversations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-User-ID': userId },
    body: JSON.stringify({ message }),
  })
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  const data = await res.json()
  const msg = typeof data.message === 'string' ? JSON.parse(data.message) : data.message
  const parsed = JSON.parse(msg.content)
  return {
    conversationId: data.conversation_id,
    reply: { id: msg.id, role: 'assistant', content: parsed.message, timestamp: new Date() },
  }
}

export async function sendMessage(userId: string, conversationId: string, message: string): Promise<ChatMessage> {
  const res = await fetch(`${API_BASE}/conversations/${conversationId}/messages`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-User-ID': userId },
    body: JSON.stringify({ message }),
  })
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  const data = await res.json()
  const msg = typeof data.message === 'string' ? JSON.parse(data.message) : data.message
  const parsed = JSON.parse(msg.content)
  return { id: msg.id, role: 'assistant', content: parsed.message, timestamp: new Date() }
}

export async function fetchHairaSource(): Promise<string> {
  const res = await fetch(`${API_BASE}/haira-source`)
  if (!res.ok) return '// Could not load source'
  const data = await res.json()
  return data.source
}
