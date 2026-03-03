# Cloudflare AI Agent — Deployment Guide

## Overview

This PoC demonstrates Haira's `backend: "cloudflare"` provider with automatic endpoint resolution. The agent uses Cloudflare Workers AI (Llama 3.1 70B) with web search, weather, calculator, and time tools.

## Prerequisites

- Haira compiler installed (`make install`)
- Cloudflare account with Workers AI enabled
- Cloudflare API token with **Workers AI** permission
- Docker (for containerized deployment)

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `CLOUDFLARE_API_KEY` | Yes | Cloudflare API token |
| `CLOUDFLARE_ACCOUNT_ID` | Yes | Cloudflare account ID (Dashboard > Workers & Pages > Account ID) |

## Local Development

```bash
# Set credentials
export CLOUDFLARE_API_KEY="your-api-token"
export CLOUDFLARE_ACCOUNT_ID="your-account-id"

# Run directly
haira run poc/cloudflare-agent/main.haira

# Or build and run
haira build poc/cloudflare-agent/main.haira -o cloudflare-agent
./cloudflare-agent
```

Open http://localhost:9010 for the web UI.

## Docker Deployment

```bash
# Build the haira compiler first
make build

# Build the Docker image (copy compiler binary for the build stage)
cp compiler/haira poc/cloudflare-agent/haira
docker build -t cloudflare-agent poc/cloudflare-agent/
rm poc/cloudflare-agent/haira

# Run
docker run -p 9010:9010 \
  -e CLOUDFLARE_API_KEY="your-api-token" \
  -e CLOUDFLARE_ACCOUNT_ID="your-account-id" \
  cloudflare-agent
```

## Deploy to Cloudflare

### Option 1: Cloudflare Containers (Beta)

Cloudflare Containers run Docker images on Cloudflare's edge network, managed through a Workers script.

1. **Build and push the image:**
   ```bash
   docker build -t cloudflare-agent poc/cloudflare-agent/
   docker tag cloudflare-agent registry.cloudflare.com/your-account/cloudflare-agent:latest
   docker push registry.cloudflare.com/your-account/cloudflare-agent:latest
   ```

2. **Create a Worker that manages the container** (wrangler.toml + Worker script handle routing to the container).

3. **Set secrets:**
   ```bash
   npx wrangler secret put CLOUDFLARE_API_KEY
   npx wrangler secret put CLOUDFLARE_ACCOUNT_ID
   ```

### Option 2: Any Docker Host (Fly.io, Railway, etc.)

Since Haira compiles to a static binary, it runs anywhere Docker runs:

```bash
# Fly.io
fly launch --image cloudflare-agent
fly secrets set CLOUDFLARE_API_KEY="..." CLOUDFLARE_ACCOUNT_ID="..."

# Railway
railway up
```

## How Backend Resolution Works

The `backend: "cloudflare"` field automatically resolves the API endpoint:

```
backend: "cloudflare" + account_id: "abc123"
  → https://api.cloudflare.com/client/v4/accounts/abc123/ai/v1
```

No manual endpoint URL needed. The Cloudflare Workers AI API is OpenAI-compatible, so all standard agent features (tool calling, streaming, memory) work out of the box.

## Available Models

Change the `model` field in the provider to use different Cloudflare Workers AI models:

- `@cf/meta/llama-3.1-70b-instruct` — Llama 3.1 70B (default)
- `@cf/meta/llama-3.1-8b-instruct` — Llama 3.1 8B (faster, cheaper)
- `@cf/mistral/mistral-7b-instruct-v0.2` — Mistral 7B
- `@cf/qwen/qwen1.5-14b-chat-awq` — Qwen 1.5 14B
- `@hf/thebloke/deepseek-coder-6.7b-instruct-awq` — DeepSeek Coder 6.7B
