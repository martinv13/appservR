---
name: browser-e2e-test
description: Launch appservR and drive it in a real headless browser to verify the admin UI and the reverse proxy actually work end to end - not just that `go test` passes. Use this whenever asked to browser-test, visually verify, or click through appservR's admin UI (apps/users/groups pages), or to confirm a proxied Shiny app is actually reachable through the reverse proxy after a code change. Also use it proactively after changes to server/, controllers/, templates/, modules/appserver/, or modules/vfsdata/ before claiming the admin UI or the proxy "works" - Go tests alone don't execute HTML templates in a real browser or prove a request really reaches a live app instance. Requires no R installation.
---

# appservR browser end-to-end test

`go test ./...` proves the Go logic is correct; it does not prove the admin
UI renders correctly in a browser, or that a request actually flows through
`AppServer.CreateProxy()` to a live app instance. This skill launches a real
appservR server with no R dependency and drives it with headless Chromium to
check both.

## What it does

1. Builds the `appservR` binary and the `modules/appserver/testdata/mockshiny`
   test helper (a small Go program standing in for
   `Rscript -e "shiny::runApp(...)"` - see that package's doc comment for
   details), copies the real `templates/`/`assets/` next to the binary, and
   launches the server in the background against a scratch SQLite DB.
2. Signs up as the first user (becomes admin automatically), logs in, and
   exercises the admin UI's create/list/view/delete flow for an app, a user,
   and a group - through the real templates, not test stubs.
3. Creates a *second*, active app with a real worker, waits for its
   mock-shiny instance to come up, then makes a request to the app's own
   path (not `/admin/*`) and checks the response really came from the
   proxied backend - the part that actually proves the reverse proxy works.
4. Tears everything down (server, spawned mock-shiny children, scratch dir).

## Running it

```bash
SKILL_DIR="$(dirname "$0")"  # or hardcode: <repo-root>/.claude/skills/browser-e2e-test
eval "$("$SKILL_DIR/scripts/setup.sh")"   # sets SCRATCH and PID in this shell
```

`setup.sh` builds everything, launches the server, polls until it responds,
and prints `SCRATCH=...` / `PID=...` — `eval`'ing its output sets both as
shell variables. If it fails, it prints the server log and exits non-zero;
don't proceed to the browser step.

Then drive it:

```bash
BASE="http://localhost:8080" SCRATCH="$SCRATCH" \
  NODE_PATH="$(npm root -g)" node "$SKILL_DIR/scripts/drive.js"
```

`drive.js` prints `OK: <assertion>` for each check as it passes, screenshots
to `$SCRATCH/screenshots/` (look at a few, especially if anything seems
off — a blank or broken-looking page is a real failure even if no assertion
literally caught it), and either `DONE` or `SCRIPT FAILED: <error>` at the
end with a non-zero exit code. It also prints any browser console errors it
captured (`CONSOLE_ERRORS: [...]`) — an empty array is expected; anything in
there (even if every assertion passed) is worth reading, since a page can
render its shell fine while JS quietly throws.

Always tear down afterward, even if the drive script failed midway —
leftover mock-shiny processes squat on ports and will confuse the *next*
run's `portspool` allocation or a completely unrelated `go test` run later
in the same session:

```bash
"$SKILL_DIR/scripts/teardown.sh" "$SCRATCH"
```

## Why headless Chromium via Playwright, not `chromium-cli`

`chromium-cli` isn't installed in this environment. Playwright is available
globally but not in this project's `node_modules`, so `NODE_PATH="$(npm root
-g)"` is required for `require('playwright')` to resolve. Launch with
`executablePath: '/opt/pw-browsers/chromium'` (already a symlink to the
actual Chrome binary — do not append a sub-path to it) and
`args: ['--no-sandbox']`.

## Gotchas (found the hard way — read before debugging a failure)

- **`apps.html` never goes network-idle.** It opens a persistent
  `EventSource("/admin/apps.json")` for live connected-user counts, which
  never closes on its own. `page.waitForLoadState('networkidle')` hangs
  forever on that page (30s Playwright timeout). Use `waitForSelector` for
  an element you actually need, or `waitForLoadState('domcontentloaded')`,
  everywhere in this flow — never `networkidle`.
- **A background server started with `cmd &` inside a Bash-tool-managed
  shell is not reliably killable by a later `kill $PID` from a *different*
  tool call.** The PID printed by `$!` can refer to a job-table entry that
  doesn't map cleanly across separate shell invocations. `setup.sh` uses
  `setsid nohup ... & disown` to detach it properly, and `teardown.sh` kills
  by matching the scratch directory's unique path in the process command
  line (`pkill -f "$SCRATCH/appservR"`) rather than trusting a stored PID —
  this is what actually works reliably. If you ever see `bind: address
  already in use` or SQLite `readonly database` errors, it means an old
  server from a previous run is still alive; find it with
  `fuser 8080/tcp` and kill it before retrying, don't just start another one
  on top.
- **An app's mock-shiny instance takes a moment to start.** `AppProxy`
  starts instances asynchronously (`go p.Rescale()`), so a request to a
  freshly-created active app's path can 404 or connection-refuse for the
  first second or so. `drive.js`'s `waitForProxiedApp` polls with retries
  for this reason — don't replace it with a single immediate request.
- **The apps-list card title is capitalized.** `buildAppsTemplateData` runs
  app names through `strings.Title`, so an app named `testapp` shows up as
  "Testapp" in the UI. Check for the capitalized form, or use
  `toLowerCase()` on the extracted text before comparing, or you'll get a
  false failure that looks like the app never appeared.
- **The auto-seeded "Sample-App" also spawns real mock-shiny instances.**
  `AppModelDB` seeds a default "Sample-App" (2 workers, active) into any
  fresh database, so `setup.sh`'s scratch DB starts that up automatically —
  it's not just whatever `drive.js` creates. `teardown.sh`'s pattern-based
  kill handles this fine, but don't be surprised to see mock-shiny PIDs
  before the driver script has created anything.
- **Signing up the same username twice fails (expected).** If a previous
  run's scratch dir/DB somehow leaked into this one (it shouldn't — each
  `setup.sh` call gets a fresh `mktemp -d` and DB), signup returns 500
  "Username already taken." That's correct behavior, not a bug; it means
  cleanup or isolation didn't happen and is a sign to check `teardown.sh`
  actually ran last time.
