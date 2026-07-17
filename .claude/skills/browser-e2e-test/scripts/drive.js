// Drives a running appservR instance with headless Chromium to verify the
// admin UI and the reverse proxy actually work, not just that `go test`
// passes. See SKILL.md for how this fits together.
//
// Env vars:
//   BASE       - e.g. http://localhost:8080 (required)
//   SCRATCH    - the setup.sh scratch dir, used to locate dummyapp (required)
//   SHOTDIR    - where to write screenshots (default: $SCRATCH/screenshots)
const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

const BASE = process.env.BASE;
const SCRATCH = process.env.SCRATCH;
if (!BASE || !SCRATCH) {
  console.error('BASE and SCRATCH env vars are required');
  process.exit(2);
}
const SHOTDIR = process.env.SHOTDIR || path.join(SCRATCH, 'screenshots');
fs.mkdirSync(SHOTDIR, { recursive: true });

let shotN = 0;
async function shot(page, name) {
  shotN++;
  const p = path.join(SHOTDIR, `${String(shotN).padStart(2, '0')}-${name}.png`);
  await page.screenshot({ path: p, fullPage: true });
  console.log('screenshot:', p);
}

function assert(cond, msg) {
  if (!cond) throw new Error('ASSERTION FAILED: ' + msg);
  console.log('OK:', msg);
}

// The mock backend's response body looks like:
//   mock-shiny-port=<port> path=<path> username=<u> displayedname=<d> appname=<a>
// Polling the app's own proxied path (not /admin/*) with plain fetch,
// rather than the admin UI, is what actually proves a request travels
// through AppServer.CreateProxy() to a live Instance - the admin pages only
// prove the app *record* exists, not that it's really running and reachable.
async function waitForProxiedApp(page, url, timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  let lastBody = '';
  let lastStatus = 0;
  while (Date.now() < deadline) {
    try {
      const resp = await page.request.get(url, { maxRedirects: 5 });
      lastStatus = resp.status();
      lastBody = await resp.text();
      if (lastStatus === 200 && lastBody.includes('mock-shiny-port=')) {
        return lastBody;
      }
    } catch (e) {
      // instance may not be listening yet
    }
    await new Promise(r => setTimeout(r, 300));
  }
  throw new Error(
    `timed out waiting for proxied app at ${url}; last status=${lastStatus} body=${lastBody.slice(0, 300)}`
  );
}

(async () => {
  const browser = await chromium.launch({
    executablePath: '/opt/pw-browsers/chromium',
    args: ['--no-sandbox'],
  });
  const page = await (await browser.newContext()).newPage();
  const consoleErrors = [];
  page.on('console', msg => {
    if (msg.type() === 'error') consoleErrors.push(msg.text());
  });
  page.on('pageerror', err => consoleErrors.push('pageerror: ' + err.message));

  // --- Sign up as the first user -> becomes admin automatically ---
  await page.goto(`${BASE}/auth/signup`);
  await shot(page, 'signup-page');
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="displayedname"]', 'Admin User');
  await page.fill('input[name="password"]', 'password123');
  await page.fill('input[name="password2"]', 'password123');
  await page.click('button[type="submit"]');
  await page.waitForLoadState('domcontentloaded');
  await shot(page, 'signup-success');

  await page.goto(`${BASE}/auth/login`);
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'password123');
  await page.click('button[type="submit"]');
  await page.waitForLoadState('domcontentloaded');
  await shot(page, 'after-login');
  assert(!page.url().includes('/auth/login'), 'login redirected away from the login page');

  // ================= Admin UI CRUD: apps, users, groups =================
  // apps.html opens a persistent EventSource("/admin/apps.json") for live
  // connected-user counts. That connection never closes on its own, so
  // page.waitForLoadState('networkidle') hangs forever on this page (and on
  // any page you navigate to next, since the old page's SSE connection can
  // linger). Use waitForSelector or 'domcontentloaded' instead - see the Go
  // skill docs for this same gotcha in Playwright's own docs.
  await page.goto(`${BASE}/admin/apps`);
  await page.waitForSelector('#apps-row');
  await shot(page, 'apps-list');

  await page.goto(`${BASE}/admin/apps/new`);
  await page.fill('#appname', 'crudtestapp');
  await page.fill('#path', '/crudtestapp');
  await page.fill('#appdir', path.join(SCRATCH, 'dummyapp'));
  // Deliberately left inactive/0 workers: this app only exists to prove the
  // CRUD flow and its own detail page render correctly, not to be proxied
  // to (see the separate active app below for that).
  await page.click('button.btn-success');
  await page.waitForLoadState('domcontentloaded');
  await shot(page, 'app-created');
  assert((await page.textContent('body')).includes('successfuly'), 'app create shows success message');

  await page.goto(`${BASE}/admin/apps/crudtestapp`);
  await shot(page, 'app-detail');
  assert((await page.textContent('body')).includes('Console output'), 'app detail page renders console output section');

  await page.goto(`${BASE}/admin/apps`);
  await page.waitForSelector('#apps-row');
  assert((await page.textContent('#apps-row')).includes('Crudtestapp'), 'apps list shows the created app (capitalized via strings.Title)');

  await page.goto(`${BASE}/admin/apps/crudtestapp`);
  await page.click('button[data-target="#delete-app-modal"]');
  await page.waitForSelector('#delete-app-modal.show');
  await shot(page, 'app-delete-modal');
  await page.click('#delete-app-modal a.btn-danger');
  await page.waitForLoadState('domcontentloaded');
  await shot(page, 'app-deleted');
  assert((await page.textContent('body')).includes('has been deleted'), 'app delete shows confirmation message');

  // Users
  await page.goto(`${BASE}/admin/users/new`);
  await page.fill('#username', 'bob');
  await page.fill('#displayedname', 'Bob Bobson');
  await page.fill('#password', 'secret123');
  await page.click('button[type="submit"]');
  await page.waitForLoadState('domcontentloaded');
  await shot(page, 'user-created');

  await page.goto(`${BASE}/admin/users`);
  await page.waitForSelector('#users-row');
  assert((await page.textContent('#users-row')).includes('Bob Bobson'), 'users list shows the created user');
  await shot(page, 'users-list');

  await page.goto(`${BASE}/admin/users/bob`);
  await page.click('button[data-target="#delete-user-modal"]');
  await page.waitForSelector('#delete-user-modal.show');
  await page.click('#delete-user-modal a.btn-danger');
  await page.waitForLoadState('domcontentloaded');
  assert((await page.textContent('body')).includes('has been deleted'), 'user delete shows confirmation message');

  // Groups
  await page.goto(`${BASE}/admin/groups/new`);
  await page.fill('#groupname', 'editors');
  await page.click('button[type="submit"]');
  await page.waitForLoadState('domcontentloaded');
  await shot(page, 'group-created');

  await page.goto(`${BASE}/admin/groups`);
  await page.waitForSelector('#groups-row');
  assert((await page.textContent('#groups-row')).includes('editors'), 'groups list shows the created group');

  await page.goto(`${BASE}/admin/groups/editors`);
  await page.click('button[data-target="#delete-group-modal"]');
  await page.waitForSelector('#delete-group-modal.show');
  await page.click('#delete-group-modal a.btn-danger');
  await page.waitForLoadState('domcontentloaded');
  assert((await page.textContent('body')).includes('has been deleted'), 'group delete shows confirmation message');

  // ================= End-to-end proxied app test =================
  // This is the part that actually proves the reverse proxy works: create
  // an ACTIVE app with a real worker, wait for its mock-shiny instance to
  // come up, then hit the app's own path (not /admin/*) and confirm the
  // response really came from the proxied backend.
  await page.goto(`${BASE}/admin/apps/new`);
  await page.fill('#appname', 'proxytestapp');
  await page.fill('#path', '/proxytestapp');
  await page.fill('#appdir', path.join(SCRATCH, 'dummyapp'));
  await page.check('#active');
  await page.fill('#workers', '1');
  await page.click('button.btn-success');
  await page.waitForLoadState('domcontentloaded');
  await shot(page, 'proxy-app-created');
  assert((await page.textContent('body')).includes('successfuly'), 'active app create shows success message');

  const proxyBody = await waitForProxiedApp(page, `${BASE}/proxytestapp/`, 15000);
  assert(proxyBody.includes('mock-shiny-port='), 'proxied request body identifies the mock backend port');
  assert(proxyBody.includes('appname=proxytestapp'), 'appservR-appname header was forwarded to the backend');
  console.log('proxied app response:', proxyBody.trim());

  // Sanity: hitting the admin-authenticated page as the SAME browser
  // session still works after all this (session/cookies survived), and the
  // detail page shows the instance's own console output - proving the
  // admin UI's live status view and the actual running process agree.
  // page.textContent() only returns rendered text nodes, never HTML
  // attribute values (an earlier version of this check looked for
  // value="proxytestapp", which can never appear in textContent - use
  // page.inputValue() for form field values instead).
  await page.goto(`${BASE}/admin/apps/proxytestapp`);
  await shot(page, 'proxy-app-detail-running');
  assert((await page.inputValue('#appname')) === 'proxytestapp', 'app detail form field still shows the right app name');
  assert((await page.textContent('body')).includes('Listening on'), 'app detail page shows the running instance\'s console output');

  await page.goto(`${BASE}/admin/apps/proxytestapp/delete`);
  await page.waitForLoadState('domcontentloaded');

  console.log('CONSOLE_ERRORS:', JSON.stringify(consoleErrors));
  console.log('DONE');
  await browser.close();
})().catch(err => {
  console.error('SCRIPT FAILED:', err);
  process.exit(1);
});
