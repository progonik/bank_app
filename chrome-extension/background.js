const BACKEND_URL = 'https://ddd.hijack.uz';
const POLL_INTERVAL_MINUTES = 20;

chrome.alarms.create('poll', { periodInMinutes: POLL_INTERVAL_MINUTES });
chrome.alarms.onAlarm.addListener(alarm => {
  if (alarm.name === 'poll') checkAndSync();
});

// Run once on service worker startup
checkAndSync();

async function checkAndSync() {
  try {
    const refreshUrl = `${BACKEND_URL}/api/v1/entrepreneurs/birdarcha-token/needs-refresh`;
    const resp = await fetch(refreshUrl).catch(err => {
      throw new Error(`needs-refresh network error for ${refreshUrl}: ${err.name}: ${err.message}`);
    });
    if (!resp.ok) {
      const body = await resp.text().catch(() => '');
      throw new Error(`needs-refresh failed: ${resp.status} ${resp.statusText}: ${body}`);
    }

    const { needs_refresh } = await resp.json();
    if (!needs_refresh) return;

    const tabs = await chrome.tabs.query({ url: '*://*.birdarcha.uz/*' });
    if (tabs.length === 0) return;

    const tab = tabs[0];

    await chrome.tabs.reload(tab.id);
    await waitForTabLoad(tab.id);
    await sleep(2000); // wait for the JS app to rehydrate localStorage

    const results = await chrome.scripting.executeScript({
      target: { tabId: tab.id },
      func: extractToken,
    });

    const token = results?.[0]?.result;
    if (!token) throw new Error('persist:auth accessToken not found after Birdarcha reload');

    const updateUrl = `${BACKEND_URL}/api/v1/entrepreneurs/birdarcha-token`;
    const updateResp = await fetch(updateUrl, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token }),
    }).catch(err => {
      throw new Error(`token update network error for ${updateUrl}: ${err.name}: ${err.message}`);
    });
    if (!updateResp.ok) {
      const body = await updateResp.text().catch(() => '');
      throw new Error(`token update failed: ${updateResp.status} ${updateResp.statusText}: ${body}`);
    }
  } catch (err) {
    console.error('birdarcha auto-sync error:', err);
  }
}

function extractToken() {
  try {
    const raw = localStorage.getItem('persist:auth');
    if (!raw) return null;
    const outer = JSON.parse(raw);
    if (!outer.tokens) return null;
    const inner = JSON.parse(outer.tokens);
    return inner.accessToken || null;
  } catch {
    return null;
  }
}

async function waitForTabLoad(tabId, timeout = 15000) {
  const start = Date.now();
  while (Date.now() - start < timeout) {
    const tab = await chrome.tabs.get(tabId);
    if (tab.status === 'complete') return;
    await sleep(300);
  }
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
