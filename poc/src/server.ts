// Haira Runtime — HTTP server (Bun.serve) matching the frontend API contract

import type { AgentConfig } from './parser';
import { runAgent, findSessionByConversation, listSessions, deleteSession } from './agent';
import { randomUUID } from 'crypto';

let agentConfig: AgentConfig;

export function initServer(agent: AgentConfig) {
  agentConfig = agent;
}

function corsHeaders(): Record<string, string> {
  return {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type, X-User-ID',
  };
}

function jsonResponse(data: any, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: { 'Content-Type': 'application/json', ...corsHeaders() },
  });
}

export function startServer(port: number) {
  const server = Bun.serve({
    port,
    async fetch(req) {
      const url = new URL(req.url);
      const path = url.pathname;

      // CORS preflight
      if (req.method === 'OPTIONS') {
        return new Response(null, { status: 204, headers: corsHeaders() });
      }

      // Serve main.haira source for the frontend code view
      if (path === '/api/haira-source' && req.method === 'GET') {
        try {
          const source = await Bun.file('./main.haira').text();
          return jsonResponse({ source });
        } catch {
          return jsonResponse({ error: 'Could not read main.haira' }, 500);
        }
      }

      // POST /api/conversations — Create new conversation
      if (path === '/api/conversations' && req.method === 'POST') {
        const userId = req.headers.get('X-User-ID') || 'anonymous';
        const body = await req.json();
        const message = body.message as string;

        if (!message) {
          return jsonResponse({ error: 'message is required' }, 400);
        }

        const conversationId = randomUUID();
        const sessionId = `${userId}:${conversationId}`;

        console.log(`[new conversation] ${conversationId} from ${userId}`);
        console.log(`[user] ${message}`);

        try {
          const reply = await runAgent(agentConfig, message, sessionId, conversationId);
          console.log(`[assistant] ${reply.substring(0, 100)}...`);

          return jsonResponse({
            conversation_id: conversationId,
            message: {
              id: randomUUID(),
              role: 'assistant',
              content: JSON.stringify({ message: reply, data: null }),
            },
          });
        } catch (err: any) {
          console.error(`[error] ${err.message}`);
          return jsonResponse({
            conversation_id: conversationId,
            message: {
              id: randomUUID(),
              role: 'assistant',
              content: JSON.stringify({ message: 'Sorry, an error occurred. Please try again.', data: null }),
            },
          });
        }
      }

      // POST /api/conversations/:id/messages — Send message
      const msgMatch = path.match(/^\/api\/conversations\/([^/]+)\/messages$/);
      if (msgMatch && req.method === 'POST') {
        const conversationId = msgMatch[1];
        const userId = req.headers.get('X-User-ID') || 'anonymous';
        const body = await req.json();
        const message = body.message as string;

        if (!message) {
          return jsonResponse({ error: 'message is required' }, 400);
        }

        const sessionId = `${userId}:${conversationId}`;
        console.log(`[message] ${conversationId}: ${message}`);

        try {
          const reply = await runAgent(agentConfig, message, sessionId, conversationId);
          console.log(`[assistant] ${reply.substring(0, 100)}...`);

          return jsonResponse({
            message: {
              id: randomUUID(),
              role: 'assistant',
              content: JSON.stringify({ message: reply, data: null }),
            },
          });
        } catch (err: any) {
          console.error(`[error] ${err.message}`);
          return jsonResponse({
            message: {
              id: randomUUID(),
              role: 'assistant',
              content: JSON.stringify({ message: 'Sorry, an error occurred.', data: null }),
            },
          });
        }
      }

      // GET /api/conversations — List conversations
      if (path === '/api/conversations' && req.method === 'GET') {
        const allSessions = listSessions();
        return jsonResponse({
          conversations: allSessions.map(s => ({
            id: s.conversationId,
            title: 'Craftify Chat',
            preview: '',
            message_count: s.messageCount,
            last_updated: s.createdAt.toISOString(),
          })),
        });
      }

      // GET /api/conversations/:id — Get conversation detail
      const getMatch = path.match(/^\/api\/conversations\/([^/]+)$/);
      if (getMatch && req.method === 'GET') {
        const session = findSessionByConversation(getMatch[1]);
        if (!session) {
          return jsonResponse({ id: getMatch[1], messages: [] });
        }
        return jsonResponse({
          id: getMatch[1],
          messages: session.messages
            .filter(m => m.role === 'user' || m.role === 'assistant')
            .map(m => ({
              id: randomUUID(),
              role: m.role,
              content: typeof m.content === 'string' ? m.content : '',
            })),
        });
      }

      // DELETE /api/conversations/:id
      const delMatch = path.match(/^\/api\/conversations\/([^/]+)$/);
      if (delMatch && req.method === 'DELETE') {
        deleteSession(delMatch[1]);
        return new Response(null, { status: 204, headers: corsHeaders() });
      }

      return jsonResponse({ error: 'Not found' }, 404);
    },
  });

  console.log(`Craftify running on :${server.port}`);
  return server;
}
