// Haira Runtime — Agent execution engine (Azure OpenAI with tool calling)

import { AzureOpenAI } from 'openai';
import type { ChatCompletionMessageParam } from 'openai/resources/chat/completions';
import { toolRegistry, toolDefinitions } from './tools';
import type { AgentConfig } from './parser';

interface Session {
  messages: ChatCompletionMessageParam[];
  conversationId: string;
  createdAt: Date;
}

const sessions = new Map<string, Session>();

export let openaiClient: AzureOpenAI;

export function initAgent(client: AzureOpenAI) {
  openaiClient = client;
}

export function getOrCreateSession(sessionId: string, conversationId: string): Session {
  let session = sessions.get(sessionId);
  if (!session) {
    session = { messages: [], conversationId, createdAt: new Date() };
    sessions.set(sessionId, session);
  }
  return session;
}

export function getSession(sessionId: string): Session | undefined {
  return sessions.get(sessionId);
}

export function listSessions(): { id: string; conversationId: string; messageCount: number; createdAt: Date }[] {
  const result: { id: string; conversationId: string; messageCount: number; createdAt: Date }[] = [];
  for (const [id, session] of sessions) {
    result.push({
      id,
      conversationId: session.conversationId,
      messageCount: session.messages.filter(m => m.role === 'user' || m.role === 'assistant').length,
      createdAt: session.createdAt,
    });
  }
  return result;
}

export function deleteSession(conversationId: string): boolean {
  for (const [id, session] of sessions) {
    if (session.conversationId === conversationId) {
      sessions.delete(id);
      return true;
    }
  }
  return false;
}

export function findSessionByConversation(conversationId: string): Session | undefined {
  for (const session of sessions.values()) {
    if (session.conversationId === conversationId) return session;
  }
  return undefined;
}

export async function runAgent(
  agent: AgentConfig,
  message: string,
  sessionId: string,
  conversationId: string,
): Promise<string> {
  const session = getOrCreateSession(sessionId, conversationId);

  // Add user message
  session.messages.push({ role: 'user', content: message });

  // Trim to max turns
  const maxMessages = agent.memoryMaxTurns * 2; // 2 messages per turn (user + assistant)
  if (session.messages.length > maxMessages) {
    session.messages = session.messages.slice(-maxMessages);
  }

  // Build full messages array with system prompt
  const fullMessages: ChatCompletionMessageParam[] = [
    { role: 'system', content: agent.system },
    ...session.messages,
  ];

  // Filter tool definitions to only include agent's tools
  const agentTools = toolDefinitions.filter(t =>
    agent.tools.includes(t.function.name)
  );

  // Agent loop — keep calling LLM until we get a text response
  const MAX_ITERATIONS = 10;
  for (let i = 0; i < MAX_ITERATIONS; i++) {
    console.log(`  [agent] iteration ${i + 1}, messages: ${fullMessages.length}`);

    const response = await openaiClient.chat.completions.create({
      model: process.env.AZURE_OPENAI_DEPLOYMENT_NAME || 'gpt-4o',
      messages: fullMessages,
      tools: agentTools.length > 0 ? agentTools : undefined,
      temperature: agent.temperature,
    });

    const choice = response.choices[0];

    if (choice.finish_reason === 'tool_calls' || choice.message.tool_calls?.length) {
      // LLM wants to call tools
      fullMessages.push(choice.message);

      for (const toolCall of choice.message.tool_calls || []) {
        const toolName = toolCall.function.name;
        const toolArgs = JSON.parse(toolCall.function.arguments);

        console.log(`  [tool] calling ${toolName}(${JSON.stringify(toolArgs)})`);

        const toolFn = toolRegistry[toolName];
        let toolResult: string;
        if (toolFn) {
          try {
            toolResult = await toolFn(toolArgs);
          } catch (err: any) {
            toolResult = JSON.stringify({ error: err.message });
          }
        } else {
          toolResult = JSON.stringify({ error: `Unknown tool: ${toolName}` });
        }

        console.log(`  [tool] ${toolName} returned ${toolResult.length} chars`);

        fullMessages.push({
          role: 'tool',
          tool_call_id: toolCall.id,
          content: toolResult,
        });
      }
    } else {
      // LLM returned a text response — we're done
      const content = choice.message.content || '';
      session.messages.push({ role: 'assistant', content });
      return content;
    }
  }

  const fallback = "I'm sorry, I couldn't complete that request. Please try again.";
  session.messages.push({ role: 'assistant', content: fallback });
  return fallback;
}
