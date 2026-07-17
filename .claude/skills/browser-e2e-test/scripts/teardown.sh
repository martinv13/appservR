#!/usr/bin/env bash
# Stops the server and any mock-shiny app instances it spawned, then
# removes the scratch directory. Safe to call even if setup.sh partially
# failed or the server already died.
#
# Usage: teardown.sh <scratch-dir> [port]
set -uo pipefail

SCRATCH="${1:?usage: teardown.sh <scratch-dir> [port]}"
PORT="${2:-8080}"

# Killing by pattern (scoped to this run's unique scratch path) catches the
# server AND every mock-shiny instance it started as a child, even though
# they were never tracked by this shell's job table (see setup.sh's
# setsid/disown comment - a plain child PID is not reliable here).
#
# One pkill pass isn't always enough: deleting/deactivating an app kicks
# off an async AppProxy.Rescale()/Instance.Stop() on the server side (see
# modules/appserver/app.go's `go p.Rescale()`), which can still be spawning
# or stopping a mock-shiny child in a narrow window around when the parent
# server process gets killed - a forked-but-not-yet-exec'd child can be
# invisible to a pattern match for a moment, then show up as "mockshiny"
# a beat later, orphaned. So repeat kill+check rather than kill once and
# just wait.
remaining=""
for i in 1 2 3 4 5 6; do
  pkill -9 -f "$SCRATCH/appservR" 2>/dev/null
  pkill -9 -f "$SCRATCH/mockshiny" 2>/dev/null
  fuser -k "${PORT}/tcp" 2>/dev/null
  sleep 0.5
  remaining="$(pgrep -f "$SCRATCH" 2>/dev/null || true)"
  [ -z "$remaining" ] && break
done

rm -rf "$SCRATCH"

if [ -n "$remaining" ]; then
  echo "WARNING: processes still referencing $SCRATCH: $remaining" >&2
fi
