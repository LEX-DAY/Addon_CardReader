const HOST_NAME = "com.cardreader.bridge";

let port = null;

function ensurePort() {
  if (port) {
    return;
  }

  port = chrome.runtime.connectNative(HOST_NAME);

  port.onMessage.addListener((msg) => {
    broadcastToTabs(msg);
  });

  port.onDisconnect.addListener(() => {
    const err = chrome.runtime.lastError;
    console.error("Native host disconnected", err?.message ?? "");
    port = null;

    setTimeout(() => {
      ensurePort();
    }, 1500);
  });
}

function broadcastToTabs(msg) {
  chrome.tabs.query({}, (tabs) => {
    for (const tab of tabs) {
      if (!tab.id) continue;
      chrome.tabs.sendMessage(tab.id, {
        type: "CARD_READER_DATA",
        payload: msg,
      });
    }
  });
}

chrome.runtime.onInstalled.addListener(() => {
  ensurePort();
});

chrome.runtime.onStartup.addListener(() => {
  ensurePort();
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg?.type !== "CARD_READER_SEND" || !msg.payload) {
    return;
  }

  ensurePort();
  port.postMessage(msg.payload);
  sendResponse({ ok: true });
});

ensurePort();
