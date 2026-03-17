#!/usr/bin/with-contenv bash
# Enables Obsidian community plugins headlessly via Chrome DevTools Protocol.
#
# Obsidian 1.x shows a "Do you trust the author?" dialog on first vault open
# which blocks plugin loading in headless environments. This script bypasses
# it by calling app.plugins.setEnable(true) directly.
#
# Requires Obsidian started with:
#   --remote-debugging-port=9222 --remote-allow-origins=*

CDP_URL="http://127.0.0.1:9222"
REST_API="http://127.0.0.1:27124"
MAX_WAIT=90
WAITED=0

echo "**** auto-trust: waiting for CDP ****"

while [ "$WAITED" -lt "$MAX_WAIT" ]; do
    if curl -sf "${CDP_URL}/json" >/dev/null 2>&1; then
        break
    fi
    sleep 2
    WAITED=$((WAITED + 2))
done

if [ "$WAITED" -ge "$MAX_WAIT" ]; then
    echo "**** auto-trust: timed out waiting for CDP ****"
    exit 1
fi

# Give the vault page time to load
sleep 8

WS_URL=$(curl -sf "${CDP_URL}/json" | python3 -c "
import sys, json
for p in json.load(sys.stdin):
    if 'index.html' in p.get('url', ''):
        print(p['webSocketDebuggerUrl'])
        break
" 2>/dev/null)

if [ -z "$WS_URL" ]; then
    echo "**** auto-trust: vault page not found ****"
    exit 1
fi

export WS_URL
python3 << 'PYEOF'
import json, os, sys, time

try:
    import websocket
except ImportError:
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "-q", "websocket-client"])
    import websocket

ws = websocket.create_connection(os.environ["WS_URL"])
seq = [0]

def cdp_eval(expr):
    seq[0] += 1
    ws.send(json.dumps({
        "id": seq[0],
        "method": "Runtime.evaluate",
        "params": {"expression": expr, "awaitPromise": True},
    }))
    return json.loads(ws.recv())

# Wait for the Obsidian app object
for _ in range(20):
    r = cdp_eval("typeof app !== 'undefined' && !!app.vault")
    if r.get("result", {}).get("result", {}).get("value") is True:
        break
    time.sleep(2)
else:
    print("**** auto-trust: app not ready ****", file=sys.stderr)
    ws.close()
    sys.exit(1)

# Dismiss the trust dialog if present
cdp_eval("""
(() => {
    const btn = document.querySelector('.modal-button-container button');
    if (btn) btn.click();
})()
""")
time.sleep(1)

# Enable community plugins
r = cdp_eval("(async()=>{await app.plugins.setEnable(true);return 'ok'})()")
val = r.get("result", {}).get("result", {}).get("value", "?")
print(f"**** auto-trust: setEnable → {val} ****")

ws.close()
PYEOF

echo "**** auto-trust: waiting for REST API ****"
for _ in $(seq 1 15); do
    if curl -sf "$REST_API" >/dev/null 2>&1; then
        echo "**** auto-trust: REST API is up ****"
        exit 0
    fi
    sleep 2
done

echo "**** auto-trust: REST API did not start ****"
exit 1
