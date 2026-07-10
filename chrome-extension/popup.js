const DEFAULT_STORAGE_KEY = 'persist:auth';
const BACKEND_URL = 'https://ddd.hijack.uz';

const $ = id => document.getElementById(id);

function setStatus(state, text) {
  const dot = $('status-dot');
  dot.className = `dot ${state}`;
  const statusText = $('status-text');
  statusText.textContent = text;
  statusText.className = state === 'error' ? 'status-text error-text' : 'status-text';
}

function setErrorDetail(text) {
  const el = $('error-detail');
  if (!text) {
    el.style.display = 'none';
    el.textContent = '';
    return;
  }
  el.style.display = 'block';
  el.textContent = text;
}

function showTokenPreview(token) {
  const el = $('token-preview');
  if (!token) { el.style.display = 'none'; return; }
  el.style.display = 'block';
  el.textContent = token.length > 80 ? token.slice(0, 80) + '…' : token;
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
  const birdarchaToken = await readTokenFromPage(DEFAULT_STORAGE_KEY);
  showTokenPreview(birdarchaToken);

  const url = `${BACKEND_URL}/api/v1/entrepreneurs/birdarcha-token`;
  let resp;
  try {
    resp = await fetch(url, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token: birdarchaToken }),
    });
  } catch (err) {
    const reason = err instanceof Error ? `${err.name}: ${err.message}` : String(err);
    throw new Error(
      [
        'Network request failed.',
        `URL: ${url}`,
        `Reason: ${reason}`,
        '',
        'Common causes:',
        '- backend domain does not resolve',
        '- internet/VPN/proxy blocked the request',
        '- extension is outdated and points to the wrong backend',
      ].join('\n')
    );
  }

  if (!resp.ok) {
    const body = await resp.text().catch(() => '');
    throw new Error(
      [
        'Backend rejected token update.',
        `URL: ${url}`,
        `Status: ${resp.status} ${resp.statusText}`,
        `Body: ${body || '<empty>'}`,
      ].join('\n')
    );
  }
}

async function onSyncClick() {
  $('sync-btn').disabled = true;
  setStatus('loading', 'Syncing…');
  setErrorDetail(null);
  showTokenPreview(null);

  try {
    await syncToken();
    setStatus('ok', 'Token synced successfully');
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    const firstLine = message.split('\n')[0] || 'Sync failed';
    setStatus('error', firstLine);
    setErrorDetail(message);
  } finally {
    $('sync-btn').disabled = false;
  }
}

document.addEventListener('DOMContentLoaded', () => {
  $('sync-btn').addEventListener('click', onSyncClick);
});
