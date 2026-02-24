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
    "input:not([type='hidden']), textarea, [contenteditable='true']"
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
    return;
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
    return;
  }

  if (target.isContentEditable) {
    target.textContent = text;
    target.dispatchEvent(new Event("input", { bubbles: true }));
    target.dispatchEvent(new Event("change", { bubbles: true }));
  }
}

chrome.runtime.onMessage.addListener((message) => {
  if (message?.type !== "CARD_READER_DATA") {
    return;
  }

  const payload = message.payload;
  const formatted = payload?.w34b?.expandedHex || payload?.w26?.cardNumber || payload?.raw || "";
  if (!formatted) {
    return;
  }

  injectValue(String(formatted));
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
