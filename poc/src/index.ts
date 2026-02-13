// Haira Runtime — Entry point
// Reads a .haira file, parses agentic constructs, and starts the server

import { Pool } from "pg";
import { AzureOpenAI } from "openai";
import { parseHairaFile } from "./parser";
import { initTools } from "./tools";
import { initAgent } from "./agent";
import { initServer, startServer } from "./server";

async function main() {
  // Read the .haira source file
  // When invoked from `haira serve`, HAIRA_SOURCE_FILE is set and parsing was already done
  const fromCompiler = !!process.env.HAIRA_SOURCE_FILE;
  const hairaPath =
    process.env.HAIRA_SOURCE_FILE || process.argv[2] || "./main.haira";

  if (!fromCompiler) {
    console.log(`[haira] Reading ${hairaPath}...`);
  }

  const source = await Bun.file(hairaPath).text();
  const program = parseHairaFile(source);

  if (!fromCompiler) {
    console.log(`[haira] Parsed:`);
    console.log(
      `  ${program.providers.length} provider(s): ${program.providers.map((p) => p.name).join(", ")}`,
    );
    console.log(
      `  ${program.tools.length} tool(s): ${program.tools.map((t) => t.name).join(", ")}`,
    );
    console.log(
      `  ${program.agents.length} agent(s): ${program.agents.map((a) => a.name).join(", ")}`,
    );
    console.log(
      `  ${program.workflows.length} workflow(s): ${program.workflows.map((w) => w.name).join(", ")}`,
    );
  }

  // Initialize PostgreSQL connection
  const pool = new Pool({
    host: process.env.PGHOST || "localhost",
    port: parseInt(process.env.PGPORT || "55554"),
    database: process.env.PGDATABASE || "agent-db",
    user: process.env.PGUSER || "agent-user",
    password: process.env.PGPASSWORD || "password",
  });

  // Test DB connection
  try {
    const res = await pool.query("SELECT count(*) as c FROM recipes");
    const maltRes = await pool.query("SELECT count(*) as c FROM malts");
    console.log(
      `[runtime] pgvector connected — ${res.rows[0].c} recipes, ${maltRes.rows[0].c} malts`,
    );
  } catch (err: any) {
    console.error(`[runtime] DB connection failed: ${err.message}`);
    process.exit(1);
  }

  // Initialize Azure OpenAI client
  const provider = program.providers[0];
  if (!provider) {
    console.error("[haira] No provider found in .haira file");
    process.exit(1);
  }

  const openai = new AzureOpenAI({
    endpoint: provider.fields.endpoint,
    apiKey: provider.fields.api_key,
    apiVersion: process.env.AZURE_OPENAI_API_VERSION || "2025-01-01-preview",
  });

  console.log(
    `[runtime] provider ${provider.name}: ${provider.fields.model} at ${provider.fields.endpoint}`,
  );

  // Initialize tools and agent
  initTools(pool, openai);
  initAgent(openai);

  // Get the agent config
  const agent = program.agents[0];
  if (!agent) {
    console.error("[haira] No agent found in .haira file");
    process.exit(1);
  }

  console.log(
    `[runtime] agent ${agent.name}: ${agent.tools.length} tools, temp=${agent.temperature}`,
  );

  // Initialize and start server
  initServer(agent);
  startServer(3000);
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
