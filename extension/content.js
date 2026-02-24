function pickTargetInput() {
  const active = document.activeElement;
  if (active instanceof HTMLInputElement || active instanceof HTMLTextAreaElement) {
    return active;
  }

  return document.querySelector("input:not([type='hidden']), textarea");
}

function injectValue(text) {
  const target = pickTargetInput();
  if (!target) {
    console.warn("Card Reader: поле ввода не найдено");
    return;
  }

  target.focus();
  target.value = text;
  target.dispatchEvent(new Event("input", { bubbles: true }));
  target.dispatchEvent(new Event("change", { bubbles: true }));
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
