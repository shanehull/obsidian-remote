#!/bin/with-contenv bash

# This script runs inside the container on startup
TEST_MODE="${TEST_MODE:-true}"
VAULT_PATH="/vaults"
OBSIDIAN_CONFIG_DIR="/config/.config/obsidian"
OBSIDIAN_JSON="$OBSIDIAN_CONFIG_DIR/obsidian.json"
KEY_FILE="/config/.obsidian_api_key"

# 1. API Key Automation
# If the user didn't provide a key, generate one and persist it
if [ -z "$OBSIDIAN_API_KEY" ]; then
    if [ -f "$KEY_FILE" ]; then
        OBSIDIAN_API_KEY=$(cat "$KEY_FILE")
    else
        OBSIDIAN_API_KEY=$(head /dev/urandom | tr -dc 'a-f0-9' | head -c 32)
        echo "$OBSIDIAN_API_KEY" > "$KEY_FILE"
    fi
fi
export OBSIDIAN_API_KEY

# 2. Vault Logic
mkdir -p "$VAULT_PATH"

if [ "$TEST_MODE" = "true" ]; then
    VAULT_ID_FILE="/config/.vault_id_test"
    if [ -f "$VAULT_ID_FILE" ]; then
        VAULT_ID=$(cat "$VAULT_ID_FILE")
    else
        VAULT_ID=$(head /dev/urandom | tr -dc 'a-f0-9' | head -c 16)
        echo "$VAULT_ID" > "$VAULT_ID_FILE"
    fi
    
    if [ ! -d "$VAULT_PATH/.obsidian" ]; then
        echo "**** TEST_MODE=true: Seeding isolated dummy vault (ID: $VAULT_ID) ****"
        mkdir -p "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api"
        echo '{"pluginEnabled": true, "community-plugin-v2": true, "nativeMenus": false, "useTitleBar": false}' > "$VAULT_PATH/.obsidian/app.json"
        echo '["obsidian-local-rest-api"]' > "$VAULT_PATH/.obsidian/community-plugins.json"
        
        echo "**** Downloading Local REST API plugin... ****"
        curl -L -s -o "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api/main.js" "https://github.com/coddingtonbear/obsidian-local-rest-api/releases/latest/download/main.js"
        curl -L -s -o "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api/manifest.json" "https://github.com/coddingtonbear/obsidian-local-rest-api/releases/latest/download/manifest.json"
        echo "{\"apiKey\":\"$OBSIDIAN_API_KEY\",\"bindAddress\":\"127.0.0.1\",\"port\":27123,\"enableInsecureServer\":true}" > "$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api/data.json"
        
        if [ ! -f "$VAULT_PATH/Welcome.md" ]; then
            echo "# Welcome to Test Vault\n\nThis is an isolated vault for API testing." > "$VAULT_PATH/Welcome.md"
        fi
    fi
else
    VAULT_ID_FILE="/config/.vault_id_real"
    if [ -f "$VAULT_ID_FILE" ]; then
        VAULT_ID=$(cat "$VAULT_ID_FILE")
    else
        VAULT_ID=$(head /dev/urandom | tr -dc 'a-f0-9' | head -c 16)
        echo "$VAULT_ID" > "$VAULT_ID_FILE"
    fi
    echo "**** TEST_MODE=false: Using real vault (ID: $VAULT_ID) ****"
    
    if [ ! -d "$VAULT_PATH/.git" ]; then
        if [ -n "$GIT_REPO_URL" ] && [ -n "$GITHUB_PAT" ]; then
            AUTH_URL=$(echo "$GIT_REPO_URL" | sed -E "s|git@github.com:|https://$GITHUB_PAT@github.com/|")
            echo "**** Cloning real vault... ****"
            sudo -u abc git clone "$AUTH_URL" "$VAULT_PATH"
        else
            echo "**** ERROR: TEST_MODE=false but GIT_REPO_URL or GITHUB_PAT missing! ****"
        fi
    fi

    # ENFORCE the API Key in the real vault every boot to keep environment in sync
    PLUGIN_DATA="$VAULT_PATH/.obsidian/plugins/obsidian-local-rest-api/data.json"
    if [ -f "$PLUGIN_DATA" ]; then
        # Use sed to update the apiKey while preserving other settings
        sed -i "s/\"apiKey\":\"[^\"]*\"/\"apiKey\":\"$OBSIDIAN_API_KEY\"/" "$PLUGIN_DATA"
    fi
fi

# 3. Ensure Obsidian config reflects selected vault
mkdir -p "$OBSIDIAN_CONFIG_DIR"
mkdir -p "/config/.config/openbox"

CONFIG_CONTENT="{\"vaults\":{\"$VAULT_ID\":{\"path\":\"$VAULT_PATH\",\"ts\":$(date +%s%3N),\"open\":true,\"trusted\":true}},\"lastOpenedVault\":\"$VAULT_ID\"}"
echo "$CONFIG_CONTENT" > "$OBSIDIAN_JSON"
echo "{}" > "$OBSIDIAN_CONFIG_DIR/$VAULT_ID.json"

# Fix the Obsidian wrapper
cat <<EOF > /usr/bin/obsidian
#!/bin/bash
exec /opt/obsidian/obsidian --no-sandbox "obsidian://open?path=%2Fvaults" "\$@"
EOF
chmod +x /usr/bin/obsidian

echo "xrandr --size 1280x720" > /config/.config/openbox/autostart
echo "obsidian &" >> /config/.config/openbox/autostart

lsiown -R abc:abc /vaults "$OBSIDIAN_CONFIG_DIR" "/config/.config/openbox"

echo "**** Init complete. ****"
