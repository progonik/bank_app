const DEFAULT_STORAGE_KEY = 'persist:auth';

const $ = id => document.getElementById(id);

function setStatus(state, text) {
  const dot = $('status-dot');
  dot.className = `dot ${state}`;
  $('status-text').textContent = text;
}

function showTokenPreview(token) {
  const el = $('token-preview');
  if (!token) { el.style.display = 'none'; return; }
  el.style.display = 'block';
  el.textContent = token.length > 80 ? token.slice(0, 80) + '…' : token;
}

async function loadSettings() {
  const data = await chrome.storage.local.get(['backendUrl']);
  $('backend-url').value = data.backendUrl || '';
}

async function saveSettings() {
  await chrome.storage.local.set({
    backendUrl: $('backend-url').value.trim().replace(/\/$/, ''),
  });
  setStatus('ok', 'Settings saved');
  setTimeout(() => setStatus('idle', 'Ready'), 1500);
}

async function readTokenFromPage(storageKey) {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.id) throw new Error('No active tab found');

  const results = await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    func: key => localStorage.getItem(key),
    args: [storageKey],
  });

  const raw = results?.[0]?.result;
  if (!raw) throw new Error(`Key "${storageKey}" not found in localStorage`);

  // Redux Persist double-encoded format:
  // localStorage["persist:auth"] = JSON { tokens: JSON { accessToken: "..." } }
  try {
    const outer = JSON.parse(raw);
    if (outer.tokens) {
      const inner = JSON.parse(outer.tokens);
      if (inner.accessToken) return inner.accessToken;
    }
  } catch (_) {}

  // Fallback: treat the raw value as the token itself
  return raw;
}

async function syncToken() {
  const data = await chrome.storage.local.get(['backendUrl']);
  const backendUrl = data.backendUrl;

  if (!backendUrl) throw new Error('Backend URL is not configured');

  const birdarchaToken = await readTokenFromPage(DEFAULT_STORAGE_KEY);
  showTokenPreview(birdarchaToken);

  const resp = await fetch(`${backendUrl}/api/v1/entrepreneurs/birdarcha-token`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token: birdarchaToken }),
  });

  if (!resp.ok) {
    const body = await resp.text().catch(() => '');
    throw new Error(`Server returned ${resp.status}: ${body}`);
  }
}

async function onSyncClick() {
  $('sync-btn').disabled = true;
  setStatus('loading', 'Syncing…');
  showTokenPreview(null);

  try {
    await syncToken();
    setStatus('ok', 'Token synced successfully');
  } catch (err) {
    setStatus('error', err.message);
  } finally {
    $('sync-btn').disabled = false;
  }
}

document.addEventListener('DOMContentLoaded', () => {
  loadSettings();
  $('sync-btn').addEventListener('click', onSyncClick);
  $('save-btn').addEventListener('click', saveSettings);
});
