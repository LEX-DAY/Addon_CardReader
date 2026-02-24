const HOST_NAME = "com.cardreader.bridge";
const KEEPALIVE_PORT = "card-reader-keepalive";

let port = null;

function detectBrowserForInstallHint() {
  const ua = navigator.userAgent.toLowerCase();
  if (ua.includes("yabrowser")) return "yandex";
  if (ua.includes("edg/")) return "edge";
  if (ua.includes("chromium")) return "chromium";
  return "chrome";
}

function logNativeHostHint(errorMessage) {
  const msg = String(errorMessage || "");
  if (!msg.toLowerCase().includes("host not found")) {
    return;
  }
  const extensionId = chrome.runtime.id;
  const browser = detectBrowserForInstallHint();
  console.error(
    `Install native host once for this browser profile. Example: cardreader-host.exe --install --extension-id ${extensionId} --browser ${browser}`
  );
}

function ensurePort() {
  if (port) {
    return;
  }

  port = chrome.runtime.connectNative(HOST_NAME);

  port.onMessage.addListener((msg) => {
    console.debug("Card Reader native message:", msg);
    broadcastToTabs(msg);
  });

  port.onDisconnect.addListener(() => {
    const err = chrome.runtime.lastError;
    console.error("Native host disconnected", err?.message ?? "");
    logNativeHostHint(err?.message ?? "");
    port = null;

    setTimeout(() => {
      ensurePort();
    }, 1500);
  });
}

function broadcastToTabs(msg) {
  const send = (tabs) => {
    for (const tab of tabs) {
      if (!tab.id) continue;
      chrome.tabs.sendMessage(tab.id, {
        type: "CARD_READER_DATA",
        payload: msg,
      }, (resp) => {
        const err = chrome.runtime.lastError;
        if (err) {
          console.debug(`Card Reader tab ${tab.id}: ${err.message}`);
          return;
        }
        if (!resp?.ok) {
          console.debug(`Card Reader tab ${tab.id}:`, resp?.reason ?? "no response");
        }
      });
    }
  };

  chrome.tabs.query({ active: true }, (tabs) => {
    const webTabs = (tabs ?? []).filter(
      (tab) => tab.id && /^(https?|file):/i.test(tab.url || "")
    );
    if (!webTabs.length) {
      console.debug("Card Reader: no active web tabs to deliver message");
      return;
    }
    send(webTabs);
  });
}

chrome.runtime.onInstalled.addListener(() => {
  ensurePort();
});

chrome.runtime.onStartup.addListener(() => {
  ensurePort();
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg?.type === "CARD_READER_READY" || msg?.type === "CARD_READER_PING") {
    ensurePort();
    sendResponse({ ok: true });
    return;
  }

  if (msg?.type !== "CARD_READER_SEND" || !msg.payload) {
    return;
  }

  ensurePort();
  port.postMessage(msg.payload);
  sendResponse({ ok: true });
});

chrome.runtime.onConnect.addListener((clientPort) => {
  if (clientPort.name !== KEEPALIVE_PORT) {
    return;
  }

  ensurePort();

  clientPort.onMessage.addListener((msg) => {
    if (msg?.type === "CARD_READER_PING") {
      ensurePort();
    }
  });
});

ensurePort();
