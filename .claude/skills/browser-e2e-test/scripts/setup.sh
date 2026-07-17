#!/usr/bin/env bash
# Builds appservR, wires it up to run without a real R installation, and
# launches it in the background. Prints two lines on success:
#   SCRATCH=<dir>
#   PID=<server pid>
# The caller (SKILL.md) uses these to drive the browser test and to tear
# down afterward via teardown.sh.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
SCRATCH="$(mktemp -d /tmp/appservr-e2e.XXXXXX)"
PORT="${APPSERVR_E2E_PORT:-8080}"

cp -r "$REPO_ROOT/templates" "$SCRATCH/"
cp -r "$REPO_ROOT/assets" "$SCRATCH/"

( cd "$REPO_ROOT" && go build -o "$SCRATCH/appservR" . )
( cd "$REPO_ROOT" && go build -o "$SCRATCH/mockshiny" ./modules/appserver/testdata/mockshiny )

mkdir -p "$SCRATCH/dummyapp"
echo "# dummy app for browser e2e test" > "$SCRATCH/dummyapp/app.R"

cat > "$SCRATCH/config.yml" <<EOF
mode: debug
rscript: $SCRATCH/mockshiny
server:
  host: localhost
  port: $PORT
database:
  type: sqlite
  path: $SCRATCH/data.db
EOF

# setsid detaches the server from this shell's job-control session, and
# disown drops it from the invoking shell's job table - without both, a
# plain "cmd &" can leave the process attached in a way that a later `kill`
# from a *different* Bash tool call won't reach, and it's easy to end up
# with two overlapping servers fighting over the same port and sqlite file
# (this happened during manual testing: a "killed" server was still very
# much alive and serving requests).
cd "$SCRATCH"
setsid nohup ./appservR serve > "$SCRATCH/server.log" 2>&1 < /dev/null &
disown
SERVER_PID=$!

for i in $(seq 1 30); do
  curl -sf "http://localhost:$PORT/auth/login" > /dev/null 2>&1 && break
  sleep 0.5
done
if ! curl -sf "http://localhost:$PORT/auth/login" > /dev/null 2>&1; then
  echo "FAILED to reach appservR on port $PORT after 15s. Log:" >&2
  cat "$SCRATCH/server.log" >&2
  exit 1
fi

echo "SCRATCH=$SCRATCH"
echo "PID=$SERVER_PID"
