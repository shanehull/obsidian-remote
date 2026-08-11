#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

tmp=$(mktemp -d)
SSE_LOG="$tmp/sse"
SSE_PID=""

cleanup() {
  echo "=== Cleaning up ==="
  kill "$SSE_PID" 2>/dev/null || true
  rm -rf "$tmp"
  docker compose down 2>&1 || true
}
trap cleanup EXIT

echo "=== Building container ==="
docker compose build 2>&1

echo "=== Starting container ==="
docker compose up -d 2>&1

# ------------------------------------------------------------------
# Wait for auto-trust and MCP bridge
# ------------------------------------------------------------------
echo "=== Waiting for startup (120s) ==="
started=$(date +%s)
timeout=120
auto_trust_ok=false
mcp_ok=false

while true; do
  elapsed=$(($(date +%s) - started))
  [ "$elapsed" -gt "$timeout" ] \
    && { echo "FAIL: timed out"; docker compose logs --tail=30; exit 1; }

  logs=$(docker compose logs --tail=50 2>&1)

  echo "$logs" | grep -q "SIGSEGV" \
    && { echo "FAIL: Obsidian crashed"; echo "$logs"; exit 1; }

  echo "$logs" | grep -q "REST API is up" && auto_trust_ok=true

  if [ "$mcp_ok" = false ]; then
    sse=$(curl -s --max-time 3 http://localhost:4000/sse 2>/dev/null || true)
    echo "$sse" | grep -q "sessionid" && mcp_ok=true
  fi

  [ "$auto_trust_ok" = true ] && [ "$mcp_ok" = true ] \
    && { echo "PASS: auto-trust done, MCP bridge up"; break; }

  sleep 3
done

# ------------------------------------------------------------------
# Test 1: Bridge health check
# ------------------------------------------------------------------
echo ""
echo "=== Test 1: Bridge /healthz ==="
code=$(curl -sf -o /dev/null -w "%{http_code}" http://localhost:4000/healthz)
[ "$code" = "200" ] && echo "PASS: /healthz $code" \
  || { echo "FAIL: /healthz $code"; exit 1; }

# ------------------------------------------------------------------
# Test 2: REST API health check
# ------------------------------------------------------------------
echo ""
echo "=== Test 2: REST API health ==="
code=$(docker compose exec -T obsidian \
  curl -sf -o /dev/null -w "%{http_code}" http://127.0.0.1:27124/)
[ "$code" = "200" ] && echo "PASS: REST API $code" \
  || { echo "FAIL: REST API $code"; exit 1; }

# ------------------------------------------------------------------
# Test 3: Create and read a note through the MCP bridge
# ------------------------------------------------------------------
echo ""
echo "=== Test 3: Create and read note via MCP bridge ==="

# Open an MCP SSE session
curl -s -N http://localhost:4000/sse > "$SSE_LOG" &
SSE_PID=$!
sleep 2
session=$(grep -o 'sessionid=[^ ]*' "$SSE_LOG" | head -1 | sed 's/sessionid=//')
[ -n "$session" ] || { echo "FAIL: no sessionId"; kill "$SSE_PID" 2>/dev/null; exit 1; }
echo "sessionId: $session"

# Complete MCP initialization handshake
echo "--- Initializing session ---"
: > "$SSE_LOG"
curl -s "http://localhost:4000/sse?sessionid=${session}" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"e2e","version":"1.0.0"}}}'
sleep 2
curl -s "http://localhost:4000/sse?sessionid=${session}" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'
sleep 1
echo "PASS: session initialized"

mcp_call() {
  local id=$1 method=$2 params=$3
  : > "$SSE_LOG"
  curl -s "http://localhost:4000/sse?sessionid=${session}" \
    -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":${id},\"method\":\"tools/call\",\"params\":${params}}"
  sleep 2
  if grep -q '"isError":true' "$SSE_LOG" || grep -q '"error"' "$SSE_LOG"; then
    echo "FAIL: ${method}"; cat "$SSE_LOG"; kill "$SSE_PID" 2>/dev/null; exit 1
  fi
}

echo "--- Creating note with sections ---"
mcp_call 1 update_note \
  '{"name":"update_note","arguments":{"path":"e2e-target.md","content":"## Details\n\nDetail content.\n\n## Notes\n\nNote content."}}'
echo "PASS: update_note succeeded"

echo "--- Reading heading via targeting ---"
mcp_call 2 read_note \
  '{"name":"read_note","arguments":{"path":"e2e-target.md","target_type":"heading","target":"Details"}}'

grep -q "Detail content" "$SSE_LOG" && echo "PASS: heading target read" \
  || { echo "FAIL: heading target mismatch"; cat "$SSE_LOG"; kill "$SSE_PID" 2>/dev/null; exit 1; }

echo "--- Reading frontmatter via targeting ---"
mcp_call 3 manage_frontmatter \
  '{"name":"manage_frontmatter","arguments":{"path":"e2e-target.md","operation":"set","jsonPayload":"{\"status\":\"done\"}"}}'
echo "PASS: set frontmatter"

mcp_call 4 read_note \
  '{"name":"read_note","arguments":{"path":"e2e-target.md","target_type":"frontmatter","target":"status"}}'

grep -q "done" "$SSE_LOG" && echo "PASS: frontmatter target read" \
  || { echo "FAIL: frontmatter target mismatch"; cat "$SSE_LOG"; kill "$SSE_PID" 2>/dev/null; exit 1; }

echo "--- Appending to heading ---"
mcp_call 5 update_note \
  '{"name":"update_note","arguments":{"path":"e2e-target.md","target_type":"heading","target":"Notes","operation":"append","content":"\n- Extra item"}}'
echo "PASS: append to heading"

mcp_call 6 read_note \
  '{"name":"read_note","arguments":{"path":"e2e-target.md","target_type":"heading","target":"Notes"}}'

grep -q "Extra item" "$SSE_LOG" && echo "PASS: appended content visible" \
  || { echo "FAIL: append not visible"; cat "$SSE_LOG"; kill "$SSE_PID" 2>/dev/null; exit 1; }

# Cleanup targeting test note
docker compose exec -T obsidian curl -sf -X DELETE \
  -H "Authorization: Bearer bridge-key" \
  http://127.0.0.1:27124/vault/e2e-target.md > /dev/null || true
echo "PASS: targeting cleanup done"

kill "$SSE_PID" 2>/dev/null || true

echo ""
echo "=== All e2e tests passed ==="
