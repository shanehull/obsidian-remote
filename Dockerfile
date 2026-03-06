FROM lscr.io/linuxserver/obsidian:latest

# Install Node.js from NodeSource
RUN apt-get update && \
    apt-get install -y curl gnupg git sudo libnss3 dbus-x11 wmctrl && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    rm -rf /var/lib/apt/lists/*

# Fix desktop entry
RUN ln -s /usr/bin/obsidian /usr/local/bin/AppRun && \
    ln -s /usr/bin/obsidian /usr/bin/AppRun

# Install MCP server globally
RUN npm install -g obsidian-mcp-server

# Copy vault init script
COPY init-vault.sh /custom-cont-init.d/init-vault.sh
RUN chmod +x /custom-cont-init.d/init-vault.sh

# Create MCP Server Service
RUN mkdir -p /etc/services.d/mcp-server
RUN echo "#!/usr/bin/with-contenv bash\n\
if [ -z \"\$OBSIDIAN_API_KEY\" ] || [ \"\$OBSIDIAN_API_KEY\" = \"YOUR_API_KEY\" ]; then\n\
    echo \"**** OBSIDIAN_API_KEY not set. MCP Server disabled. ****\"\n\
    sleep infinity\n\
fi\n\
export MCP_TRANSPORT_TYPE=http\n\
export MCP_HTTP_PORT=4000\n\
export MCP_HTTP_HOST=0.0.0.0\n\
export MCP_ALLOWED_ORIGINS=\"*\"\n\
export OBSIDIAN_BASE_URL=https://localhost:27123\n\
export OBSIDIAN_VERIFY_SSL=false\n\
export NODE_ENV=production\n\
# Ensure logs directory exists\n\
mkdir -p /usr/lib/node_modules/obsidian-mcp-server/logs\n\
chown -R abc:abc /usr/lib/node_modules/obsidian-mcp-server\n\
cd /usr/lib/node_modules/obsidian-mcp-server\n\
exec s6-setuidgid abc node dist/index.js" > /etc/services.d/mcp-server/run
RUN chmod +x /etc/services.d/mcp-server/run

EXPOSE 3000 4000 27123
