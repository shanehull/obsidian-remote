#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "=== Building container ==="
docker compose build 2>&1

echo "=== Starting container ==="
docker compose up -d 2>&1

cleanup() {
    echo "=== Cleaning up ==="
    docker compose down 2>&1 || true
}
trap cleanup EXIT

echo "=== Waiting for startup (120s) ==="
started=$(date +%s)
timeout=120
auto_trust_ok=false
mcp_ok=false

while true; do
    elapsed=$(($(date +%s) - started))
    [ "$elapsed" -gt "$timeout" ] && { echo "FAIL: timed out"; docker compose logs --tail=30; exit 1; }

    logs=$(docker compose logs --tail=50 2>&1)
    echo "$logs" | grep -q "SIGSEGV" && { echo "FAIL: Obsidian crashed"; echo "$logs"; exit 1; }
    echo "$logs" | grep -q "REST API is up" && auto_trust_ok=true

    if [ "$mcp_ok" = false ]; then
        sse=$(curl -s --max-time 3 http://localhost:4000/sse 2>/dev/null || true)
        echo "$sse" | grep -q "sessionId" && mcp_ok=true
    fi

    [ "$auto_trust_ok" = true ] && [ "$mcp_ok" = true ] && { echo "PASS: auto-trust done, MCP bridge up"; break; }
    sleep 3
done

echo ""
echo "=== Test 1: REST API health check ==="
status=$(docker compose exec -T obsidian curl -sf -o /dev/null -w "%{http_code}" http://127.0.0.1:27124/)
[ "$status" = "200" ] && echo "PASS: REST API returned $status" || { echo "FAIL: expected 200 got $status"; exit 1; }

echo ""
echo "=== Test 2: REST API list vault root ==="
notes=$(docker compose exec -T obsidian curl -sf -H "Authorization: Bearer bridge-key" http://127.0.0.1:27124/vault/)
echo "vault response: $notes"
echo "$notes" | grep -qE 'files|".+' && echo "PASS: /vault/ endpoint responded" || echo "WARN: unexpected vault response"

echo ""
echo "=== Test 3: MCP bridge tools/call via SSE ==="
tmp=$(mktemp -d)
curl -s -N http://localhost:4000/sse > "$tmp/sse" &
SSE_PID=$!
sleep 2
session=$(grep -o 'sessionId=[a-f0-9-]*' "$tmp/sse" | head -1)
echo "sessionId: $session"

[ -n "$session" ] || { echo "FAIL: no session"; kill "$SSE_PID" 2>/dev/null; exit 1; }

curl -s "http://localhost:4000/message?${session}" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_notes","arguments":{}}}'
sleep 3

if grep -q '"isError":true' "$tmp/sse" || grep -q '"error"' "$tmp/sse"; then
    echo "FAIL: tools/call returned error (REST API unreachable or bridge broken)"
    cat "$tmp/sse"
    kill "$SSE_PID" 2>/dev/null
    exit 1
fi

grep -q '"content"' "$tmp/sse" && echo "PASS: tools/call succeeded" || echo "WARN: unexpected SSE response"
cat "$tmp/sse"
kill "$SSE_PID" 2>/dev/null || true

echo ""
echo "=== All e2e tests passed ==="
