# Cloudflare AI Agent — Multi-stage build
#
# Haira compiles to a static Go binary — no runtime dependencies needed.
# The final image is scratch (zero base) for minimal attack surface and size.
#
# Build:  docker build -t cloudflare-agent .
# Run:    docker run -p 9010:9010 \
#           -e CLOUDFLARE_API_KEY=your-key \
#           -e CLOUDFLARE_ACCOUNT_ID=your-account-id \
#           cloudflare-agent

# --- Stage 1: Build with Haira compiler ---
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

# Install haira compiler (assumes it's published or built locally)
# For local builds, copy the binary in instead:
#   COPY haira /usr/local/bin/haira
COPY haira /usr/local/bin/haira

WORKDIR /app
COPY main.haira .

# Compile to static binary
RUN CGO_ENABLED=0 haira build main.haira -o agent

# --- Stage 2: Minimal runtime ---
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/agent /agent

EXPOSE 9010

ENTRYPOINT ["/agent"]
