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

echo "=== Waiting for startup (90s) ==="
started=$(date +%s)
timeout=90
mcp_ok=false
auto_trust_ok=false

while true; do
    elapsed=$(($(date +%s) - started))
    if [ "$elapsed" -gt "$timeout" ]; then
        echo "FAIL: timed out after ${timeout}s"
        docker compose logs --tail=30
        exit 1
    fi

    logs=$(docker compose logs --tail=50 2>&1)

    if echo "$logs" | grep -q "SIGSEGV"; then
        echo "FAIL: Obsidian crashed with SIGSEGV"
        echo "$logs"
        exit 1
    fi

    if echo "$logs" | grep -q "REST API is up"; then
        auto_trust_ok=true
    fi

    if [ "$mcp_ok" = false ]; then
        sse=$(curl -s --max-time 3 http://localhost:4000/sse 2>/dev/null || true)
        if echo "$sse" | grep -q "sessionId"; then
            mcp_ok=true
        fi
    fi

    if [ "$auto_trust_ok" = true ] && [ "$mcp_ok" = true ]; then
        echo "PASS: auto-trust completed, MCP bridge responding"
        break
    fi

    sleep 3
done

session_id=$(curl -s --max-time 3 http://localhost:4000/sse | head -1 | grep -o 'sessionId=[a-f0-9-]*')
if [ -n "$session_id" ]; then
    result=$(curl -s --max-time 5 -X POST "http://localhost:4000/message?${session_id}" \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' 2>&1)
    if echo "$result" | grep -q "obsidian"; then
        echo "PASS: MCP tools available"
    else
        echo "WARN: MCP session created but tools/list returned unexpected: $result"
    fi
fi

echo "=== All checks passed ==="
