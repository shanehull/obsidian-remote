#!/bin/with-contenv bash

# This script runs inside the container on startup
TEST_MODE="${TEST_MODE:-true}"
VAULT_PATH="/vaults"
OBSIDIAN_CONFIG_DIR="/config/.config/obsidian"
OBSIDIAN_JSON="$OBSIDIAN_CONFIG_DIR/obsidian.json"
VAULT_ID_FILE="/config/.vault_id"

# 1. Vault ID Consistency
OBSIDIAN_KEY="${OBSIDIAN_KEY:-bridge-key}"
if [ -f "$VAULT_ID_FILE" ]; then
    VAULT_ID=$(cat "$VAULT_ID_FILE")
else
    VAULT_ID=$(head /dev/urandom | tr -dc 'a-f0-9' | head -c 16)
    echo "$VAULT_ID" > "$VAULT_ID_FILE"
fi

# 2. Vault Sync Logic
mkdir -p "$VAULT_PATH"
if [ "$TEST_MODE" = "true" ]; then
    if [ ! -d "$VAULT_PATH/.obsidian" ]; then
        echo "**** TEST_MODE=true: Seeding isolated dummy vault ****"
        mkdir -p "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api"
        echo '{"pluginSafeMode":false}' > "$VAULT_PATH/.obsidian/app.json"
        echo '["obsidian-local-rest-api"]' > "$VAULT_PATH/.obsidian/community-plugins.json"
        # Download API plugin
        curl -L -s -o "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api/main.js" "https://github.com/coddingtonbear/obsidian-local-rest-api/releases/latest/download/main.js"
        curl -L -s -o "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api/manifest.json" "https://github.com/coddingtonbear/obsidian-local-rest-api/releases/latest/download/manifest.json"
        # The Go bridge will use its own auth logic, so we keep the internal API insecure/static
        echo "{\"apiKey\":\"$OBSIDIAN_KEY\",\"bindAddress\":\"127.0.0.1\",\"port\":27123,\"enableInsecureServer\":true,\"insecurePort\":27124}" > "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api/data.json"
    fi
else
    if [ ! -d "$VAULT_PATH/.git" ]; then
        if [ -n "$GIT_REPO_URL" ] && [ -n "$GITHUB_PAT" ]; then
            AUTH_URL=$(echo "$GIT_REPO_URL" | sed -E "s|git@github.com:|https://$GITHUB_PAT@github.com/|")
            git clone "$AUTH_URL" "$VAULT_PATH"
        fi
    else
        cd "$VAULT_PATH" && git pull
    fi
    # Override plugin config for the bridge (insecure HTTP on 27124, localhost only)
    echo "{\"apiKey\":\"$OBSIDIAN_KEY\",\"bindAddress\":\"127.0.0.1\",\"port\":27123,\"enableInsecureServer\":true,\"insecurePort\":27124}" > "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api/data.json"
fi

# 3. Ensure vault structure is complete
mkdir -p "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api"
touch "$VAULT_PATH/Welcome.md"

# Seed workspace if it doesn't exist
WORKSPACE_FILE="$VAULT_PATH/.obsidian/workspace.json"
if [ ! -f "$WORKSPACE_FILE" ]; then
    echo '{"main":{"id":"a","type":"split","children":[]},"active":"a"}' > "$WORKSPACE_FILE"
fi

# 4. AUTO-TRUST LOGIC (RFC 9728)
# Pre-seed the global obsidian.json to mark the vault as Trusted and Open
mkdir -p "$OBSIDIAN_CONFIG_DIR"
CONFIG_CONTENT="{\"vaults\":{\"$VAULT_ID\":{\"path\":\"$VAULT_PATH\",\"ts\":$(date +%s%3N),\"open\":true,\"trusted\":true}},\"lastOpenedVault\":\"$VAULT_ID\"}"
echo "$CONFIG_CONTENT" > "$OBSIDIAN_JSON"

# 4. Permissions
lsiown -R abc:abc /vaults "$OBSIDIAN_CONFIG_DIR"

echo "**** Init complete. Headless Obsidian is ready. ****"
