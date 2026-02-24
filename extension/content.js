const KEEPALIVE_PORT = "card-reader-keepalive";
const KEEPALIVE_INTERVAL_MS = 15000;
let keepalivePort = null;

function connectKeepalive() {
  if (window.top !== window) {
    return;
  }

  if (keepalivePort) {
    return;
  }

  try {
    keepalivePort = chrome.runtime.connect({ name: KEEPALIVE_PORT });
    keepalivePort.postMessage({ type: "CARD_READER_PING" });
    keepalivePort.onDisconnect.addListener(() => {
      keepalivePort = null;
      setTimeout(connectKeepalive, 1000);
    });
  } catch (_) {
    keepalivePort = null;
  }
}

function getDeepActiveElement(rootDocument = document) {
  let active = rootDocument.activeElement;
  while (active && active.shadowRoot && active.shadowRoot.activeElement) {
    active = active.shadowRoot.activeElement;
  }
  return active;
}

function pickTargetInput() {
  const active = getDeepActiveElement();
  if (active instanceof HTMLInputElement || active instanceof HTMLTextAreaElement) {
    return active;
  }

  if (active && active.isContentEditable) {
    return active;
  }

  return document.querySelector(
    "input:not([type='hidden']):not([disabled]):not([readonly]), textarea:not([disabled]):not([readonly]), [contenteditable='true']"
  );
}

function setControlValue(target, text) {
  if (target instanceof HTMLInputElement) {
    const setter = Object.getOwnPropertyDescriptor(
      HTMLInputElement.prototype,
      "value"
    )?.set;
    if (setter) {
      setter.call(target, text);
      return;
    }
  }

  if (target instanceof HTMLTextAreaElement) {
    const setter = Object.getOwnPropertyDescriptor(
      HTMLTextAreaElement.prototype,
      "value"
    )?.set;
    if (setter) {
      setter.call(target, text);
      return;
    }
  }

  target.value = text;
}

function injectValue(text) {
  const target = pickTargetInput();
  if (!target) {
    console.warn("Card Reader: input field not found");
    return { ok: false, reason: "no_target" };
  }

  if (target instanceof HTMLElement) {
    target.focus();
  }

  if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement) {
    setControlValue(target, text);
    try {
      target.dispatchEvent(
        new InputEvent("input", {
          bubbles: true,
          data: text,
          inputType: "insertText",
        })
      );
    } catch (_) {
      target.dispatchEvent(new Event("input", { bubbles: true }));
    }
    target.dispatchEvent(new Event("change", { bubbles: true }));
    return { ok: true, reason: "input_set" };
  }

  if (target.isContentEditable) {
    target.textContent = text;
    target.dispatchEvent(new Event("input", { bubbles: true }));
    target.dispatchEvent(new Event("change", { bubbles: true }));
    return { ok: true, reason: "contenteditable_set" };
  }

  return { ok: false, reason: "unsupported_target" };
}

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message?.type !== "CARD_READER_DATA") {
    return;
  }

  const payload = message.payload;
  const w26Formatted = payload?.w26
    ? `${String(payload.w26.facility).padStart(3, "0")},${payload.w26.cardNumber}`
    : "";
  const formatted = payload?.w34b?.expandedHex || w26Formatted || payload?.raw || "";
  if (!formatted) {
    sendResponse({ ok: false, reason: "empty_payload" });
    return;
  }

  try {
    sendResponse(injectValue(String(formatted)));
  } catch (e) {
    sendResponse({ ok: false, reason: e?.message || "inject_failed" });
  }
});

connectKeepalive();
chrome.runtime.sendMessage({ type: "CARD_READER_READY" });
setInterval(() => {
  connectKeepalive();
  try {
    keepalivePort?.postMessage({ type: "CARD_READER_PING" });
  } catch (_) {
    keepalivePort = null;
  }
}, KEEPALIVE_INTERVAL_MS);
