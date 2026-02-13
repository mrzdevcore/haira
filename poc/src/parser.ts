// Haira Runtime Parser — extracts agentic constructs from .haira files

export interface ProviderConfig {
  name: string;
  fields: Record<string, string>;
}

export interface ToolConfig {
  name: string;
  params: { name: string; type: string; default?: string }[];
  returnType: string;
  description: string;
}

export interface AgentConfig {
  name: string;
  model: string;
  system: string;
  tools: string[];
  temperature: number;
  memoryMaxTurns: number;
}

export interface WorkflowConfig {
  name: string;
  trigger: string;
  path: string;
}

export interface HairaProgram {
  providers: ProviderConfig[];
  tools: ToolConfig[];
  agents: AgentConfig[];
  workflows: WorkflowConfig[];
}

function resolveEnv(value: string): string {
  const envMatch = value.match(/env\("([^"]+)"\)/);
  if (envMatch) return process.env[envMatch[1]] || '';
  return value.replace(/^["']|["']$/g, '');
}

function extractTripleQuote(block: string): string {
  const match = block.match(/"""([\s\S]*?)"""/);
  return match ? match[1].trim() : '';
}

export function parseHairaFile(source: string): HairaProgram {
  const program: HairaProgram = {
    providers: [],
    tools: [],
    agents: [],
    workflows: [],
  };

  // Parse providers
  const providerRegex = /provider\s+(\w+)\s*\{([^}]+)\}/g;
  let match;
  while ((match = providerRegex.exec(source)) !== null) {
    const fields: Record<string, string> = {};
    const fieldRegex = /(\w+)\s*:\s*(.+)/g;
    let fieldMatch;
    while ((fieldMatch = fieldRegex.exec(match[2])) !== null) {
      fields[fieldMatch[1]] = resolveEnv(fieldMatch[2].trim());
    }
    program.providers.push({ name: match[1], fields });
  }

  // Parse tools — match tool blocks with balanced braces
  const toolStartRegex = /tool\s+(\w+)\s*\(([^)]*)\)\s*->\s*([^\{]+)\s*\{/g;
  while ((match = toolStartRegex.exec(source)) !== null) {
    const name = match[1];
    const paramsStr = match[2];
    const returnType = match[3].trim();

    // Find the matching closing brace
    let braceCount = 1;
    let i = match.index + match[0].length;
    while (i < source.length && braceCount > 0) {
      if (source[i] === '{') braceCount++;
      if (source[i] === '}') braceCount--;
      i++;
    }
    const toolBody = source.substring(match.index + match[0].length, i - 1);
    const description = extractTripleQuote(toolBody);

    const params = paramsStr.split(',').filter(Boolean).map(p => {
      const parts = p.trim().split(/\s*:\s*/);
      const [nameAndDefault, type] = [parts[0], parts.slice(1).join(':').trim()];
      const defaultMatch = type.match(/(.+?)\s*=\s*(.+)/);
      return {
        name: nameAndDefault,
        type: defaultMatch ? defaultMatch[1].trim() : type,
        default: defaultMatch ? defaultMatch[2].trim() : undefined,
      };
    });

    program.tools.push({ name, params, returnType, description });
  }

  // Parse agents — match agent blocks with balanced braces
  const agentStartRegex = /agent\s+(\w+)\s*\{/g;
  while ((match = agentStartRegex.exec(source)) !== null) {
    const name = match[1];
    let braceCount = 1;
    let i = match.index + match[0].length;
    while (i < source.length && braceCount > 0) {
      if (source[i] === '{') braceCount++;
      if (source[i] === '}') braceCount--;
      i++;
    }
    const agentBody = source.substring(match.index + match[0].length, i - 1);

    const modelMatch = agentBody.match(/model\s*:\s*(\w+)/);
    const system = extractTripleQuote(agentBody) ||
      (agentBody.match(/system\s*:\s*"([^"]+)"/) || [])[1] || '';
    const toolsMatch = agentBody.match(/tools\s*:\s*\[([^\]]+)\]/);
    const tempMatch = agentBody.match(/temperature\s*:\s*([\d.]+)/);
    const memoryMatch = agentBody.match(/memory\s*:\s*conversation\s*\(\s*max_turns\s*:\s*(\d+)\)/);

    program.agents.push({
      name,
      model: modelMatch ? modelMatch[1] : '',
      system,
      tools: toolsMatch ? toolsMatch[1].split(',').map(t => t.trim()) : [],
      temperature: tempMatch ? parseFloat(tempMatch[1]) : 0.7,
      memoryMaxTurns: memoryMatch ? parseInt(memoryMatch[1]) : 10,
    });
  }

  // Parse workflows
  const workflowRegex = /@(\w+)\s*\(\s*"([^"]+)"\s*\)\s*\nworkflow\s+(\w+)/g;
  while ((match = workflowRegex.exec(source)) !== null) {
    program.workflows.push({
      name: match[3],
      trigger: match[1],
      path: match[2],
    });
  }

  return program;
}
