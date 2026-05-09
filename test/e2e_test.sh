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
echo "=== Test 1: Health check ==="
status=$(docker compose exec -T obsidian curl -sf -o /dev/null -w "%{http_code}" http://127.0.0.1:27124/)
[ "$status" = "200" ] && echo "PASS: REST API returned $status" || { echo "FAIL: expected 200 got $status"; exit 1; }

echo ""
echo "=== Test 2: Create and read note via MCP bridge ==="
tmp=$(mktemp -d)
curl -s -N http://localhost:4000/sse > "$tmp/sse" &
SSE_PID=$!
sleep 2
session=$(grep -o 'sessionId=[a-f0-9-]*' "$tmp/sse" | head -1)
[ -n "$session" ] || { echo "FAIL: no sessionId"; kill "$SSE_PID" 2>/dev/null; exit 1; }
echo "sessionId: $session"

echo "--- Create via update_note ---"
: > "$tmp/sse"
curl -s "http://localhost:4000/message?${session}" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"update_note","arguments":{"path":"e2e-bridge.md","content":"# Bridge E2E\n\nCreated through MCP."}}}'
sleep 2
if grep -q '"isError":true' "$tmp/sse" || grep -q '"error"' "$tmp/sse"; then
    echo "FAIL: update_note failed"; cat "$tmp/sse"; kill "$SSE_PID" 2>/dev/null; exit 1
fi
echo "PASS: update_note succeeded"

echo "--- Read via read_note ---"
: > "$tmp/sse"
curl -s "http://localhost:4000/message?${session}" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_note","arguments":{"path":"e2e-bridge.md"}}}'
sleep 3
if grep -q '"isError":true' "$tmp/sse" || grep -q '"error"' "$tmp/sse"; then
    echo "FAIL: read_note failed"; cat "$tmp/sse"; kill "$SSE_PID" 2>/dev/null; exit 1
fi
grep -q "Bridge E2E" "$tmp/sse" && echo "PASS: content verified" || { echo "FAIL: content mismatch"; cat "$tmp/sse"; kill "$SSE_PID" 2>/dev/null; exit 1; }
cat "$tmp/sse"
kill "$SSE_PID" 2>/dev/null || true

echo "--- Cleanup ---"
docker compose exec -T obsidian curl -sf -X DELETE \
    -H "Authorization: Bearer bridge-key" \
    http://127.0.0.1:27124/vault/e2e-bridge.md > /dev/null
echo "PASS: cleanup done"

echo ""
echo "=== All e2e tests passed ==="
