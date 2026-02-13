// Haira Runtime — Tool implementations (pgvector search, cost calculation)

import { Pool } from 'pg';
import { AzureOpenAI } from 'openai';

let pool: Pool;
let openai: AzureOpenAI;

export function initTools(pgPool: Pool, aiClient: AzureOpenAI) {
  pool = pgPool;
  openai = aiClient;
}

async function embedQuery(query: string): Promise<number[]> {
  const response = await openai.embeddings.create({
    input: query,
    model: process.env.AZURE_OPENAI_EMBEDDING_DEPLOYMENT_NAME || 'text-embedding-3-small',
  });
  return response.data[0].embedding;
}

async function vectorSearch(table: string, query: string, maxResults: number): Promise<any[]> {
  const embedding = await embedQuery(query);
  const embeddingStr = `[${embedding.join(',')}]`;

  const result = await pool.query(
    `SELECT content, 1 - (embedding <=> $1::vector) AS similarity
     FROM ${table}
     ORDER BY embedding <=> $1::vector
     LIMIT $2`,
    [embeddingStr, maxResults]
  );

  return result.rows.map(row => {
    try {
      const parsed = JSON.parse(row.content);
      return { ...parsed, _similarity: Math.round(row.similarity * 100) / 100 };
    } catch {
      return { text: row.content, _similarity: Math.round(row.similarity * 100) / 100 };
    }
  });
}

export async function search_recipes(args: { query: string; max_results?: number }): Promise<string> {
  const results = await vectorSearch('recipes', args.query, args.max_results || 5);
  const simplified = results.map(r => ({
    name: r.name,
    type: r.type,
    abv: r.abv,
    ibu: r.ibu,
    tagline: r.tagline,
    batch_size: r.batch_size,
    fermentables: r.fermentables,
    hops: r.hops,
    yeasts: r.yeasts,
    similarity: r._similarity,
  }));
  return JSON.stringify(simplified, null, 2);
}

export async function search_malts(args: { query: string; max_results?: number }): Promise<string> {
  const results = await vectorSearch('malts', args.query, args.max_results || 3);
  return JSON.stringify(results, null, 2);
}

export async function calculate_cost(args: { ingredients: string; batch_size: number }): Promise<string> {
  // Estimated prices per kg for common brewing ingredients
  const prices: Record<string, number> = {
    'grain': 2.5, 'malt': 3.0, 'hop': 25.0, 'hops': 25.0,
    'yeast': 8.0, 'sugar': 1.5, 'adjunct': 4.0, 'other': 5.0,
    'extra pale': 2.8, 'pale': 2.8, 'munich': 3.2, 'crystal': 3.5,
    'chocolate': 4.0, 'roasted': 4.0, 'wheat': 2.6, 'caramalt': 3.5,
    'flaked oats': 2.5, 'dark crystal': 3.8,
  };

  let totalCost = 0;
  const breakdown: { item: string; qty: string; cost: number }[] = [];

  // Try to parse ingredients
  try {
    const ingredients = JSON.parse(args.ingredients);
    if (Array.isArray(ingredients)) {
      for (const ing of ingredients) {
        const name = (ing.name || ing.item || '').toLowerCase();
        const amount = parseFloat(ing.amount || ing.quantity || '0');
        const scaledAmount = amount * (args.batch_size / 20); // scale from default 20L
        const priceKey = Object.keys(prices).find(k => name.includes(k)) || 'other';
        const cost = scaledAmount * prices[priceKey];
        totalCost += cost;
        breakdown.push({ item: ing.name || name, qty: `${scaledAmount.toFixed(2)} kg`, cost: Math.round(cost * 100) / 100 });
      }
    }
  } catch {
    // If not JSON, estimate based on batch size
    totalCost = args.batch_size * 3.5; // rough estimate per liter
    breakdown.push({ item: 'Estimated total ingredients', qty: `${args.batch_size}L batch`, cost: totalCost });
  }

  return JSON.stringify({
    batch_size_liters: args.batch_size,
    total_ingredient_cost: Math.round(totalCost * 100) / 100,
    cost_per_liter: Math.round((totalCost / args.batch_size) * 100) / 100,
    breakdown,
    currency: 'EUR',
  }, null, 2);
}

// Registry of all tool functions
export const toolRegistry: Record<string, (args: any) => Promise<string>> = {
  search_recipes,
  search_malts,
  calculate_cost,
};

// OpenAI function definitions for tool calling
export const toolDefinitions = [
  {
    type: 'function' as const,
    function: {
      name: 'search_recipes',
      description: 'Search beer recipes by style, name, or ingredients. Uses semantic vector search to find the most relevant brewing recipes from the BrewDog DIY recipe database.',
      parameters: {
        type: 'object',
        properties: {
          query: { type: 'string', description: 'Search query for recipes (style, name, ingredients)' },
          max_results: { type: 'number', description: 'Maximum number of results to return (default: 5)' },
        },
        required: ['query'],
      },
    },
  },
  {
    type: 'function' as const,
    function: {
      name: 'search_malts',
      description: 'Find malts by their technical characteristics and flavor profiles. Uses semantic vector search on malt specifications.',
      parameters: {
        type: 'object',
        properties: {
          query: { type: 'string', description: 'Search query for malts (characteristics, flavor profile)' },
          max_results: { type: 'number', description: 'Maximum number of results to return (default: 3)' },
        },
        required: ['query'],
      },
    },
  },
  {
    type: 'function' as const,
    function: {
      name: 'calculate_cost',
      description: 'Estimate the ingredient cost for a recipe at a given batch size in liters. Returns cost breakdown per ingredient category and total.',
      parameters: {
        type: 'object',
        properties: {
          ingredients: { type: 'string', description: 'JSON array of ingredients with name and amount fields' },
          batch_size: { type: 'number', description: 'Batch size in liters' },
        },
        required: ['ingredients', 'batch_size'],
      },
    },
  },
];
