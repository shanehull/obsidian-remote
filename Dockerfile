# Stage 1: Build the Go MCP Server (Bridge)
FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -o server ./cmd/server/main.go

# Stage 2: Unified Headless Container
FROM lscr.io/linuxserver/obsidian:latest

# Install runtime dependencies for the bridge and auto-trust
RUN apt-get update && \
    apt-get install -y curl ca-certificates bash wmctrl xdotool python3-pip && \
    pip3 install --break-system-packages websocket-client && \
    rm -rf /var/lib/apt/lists/*

# Copy the Go MCP server binary from builder
COPY --from=builder /app/server /usr/local/bin/mcp-bridge

# Copy the vault init script and auto-trust helper
COPY init-vault.sh /custom-cont-init.d/init-vault.sh
COPY auto-trust.sh /usr/local/bin/auto-trust.sh
RUN chmod +x /custom-cont-init.d/init-vault.sh /usr/local/bin/auto-trust.sh

# Create the Headless Obsidian Service (with remote debugging for auto-trust)
RUN mkdir -p /etc/services.d/obsidian && \
    printf "#!/usr/bin/with-contenv bash\nexport DISPLAY=:1\nexec s6-setuidgid abc /opt/obsidian/obsidian --no-sandbox --remote-debugging-port=9222 --remote-allow-origins=* /vaults\n" > /etc/services.d/obsidian/run && \
    chmod +x /etc/services.d/obsidian/run

# Create the MCP Bridge Service
RUN mkdir -p /etc/services.d/mcp-bridge && \
    printf "#!/usr/bin/with-contenv bash\nexec s6-setuidgid abc /usr/local/bin/mcp-bridge\n" > /etc/services.d/mcp-bridge/run && \
    chmod +x /etc/services.d/mcp-bridge/run

# Auto-trust service: clicks the plugin trust dialog after Obsidian starts.
# Uses s6-svc -O to mark itself "once" so s6 doesn't restart it after success.
RUN mkdir -p /etc/services.d/auto-trust && \
    printf '#!/usr/bin/with-contenv bash\ns6-setuidgid abc /usr/local/bin/auto-trust.sh\n# Prevent s6 from restarting after completion\nsleep infinity\n' > /etc/services.d/auto-trust/run && \
    chmod +x /etc/services.d/auto-trust/run

# We only expose the MCP port. The REST API is internal.
EXPOSE 4000
