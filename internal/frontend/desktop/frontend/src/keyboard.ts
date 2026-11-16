export type KeyInput = {
  key: string;
  ctrl?: boolean;
  alt?: boolean;
  shift?: boolean;
};

const NAMED = new Set([
  "Escape",
  "Enter",
  "Backspace",
  "Tab",
  "ArrowUp",
  "ArrowDown",
  "ArrowLeft",
  "ArrowRight",
  "Home",
  "End",
  "PageUp",
  "PageDown",
  "Delete",
]);

export function attachKeyboard(ta: HTMLTextAreaElement, send: (k: KeyInput) => void) {
  let composing = false;

  const flushText = () => {
    const text = ta.value;
    ta.value = "";
    for (const ch of text) {
      send({ key: ch === "\n" ? "Enter" : ch });
    }
  };

  ta.addEventListener("keydown", (e) => {
    if (composing || e.isComposing || e.key === "Process") {
      return;
    }
    const printable = [...e.key].length === 1;
    const altGr =
      e.getModifierState("AltGraph") ||
      (e.ctrlKey && e.altKey && printable && !/^[a-z0-9]$/i.test(e.key));
    const ctrl = e.ctrlKey && !altGr;
    const alt = e.altKey && !altGr;

    if (altGr && printable) {
      e.preventDefault();
      send({ key: e.key });
      return;
    }
    if (NAMED.has(e.key) || (printable && (ctrl || alt))) {
      e.preventDefault();
      send({ key: e.key, ctrl, alt, shift: e.shiftKey });
    }
  });

  ta.addEventListener("input", (e) => {
    if (composing || (e as InputEvent).isComposing) {
      return;
    }
    flushText();
  });

  ta.addEventListener("compositionstart", () => {
    composing = true;
  });

  ta.addEventListener("compositionend", () => {
    composing = false;
    flushText();
  });

  const focus = () => ta.focus({ preventScroll: true });
  window.addEventListener("mousedown", (e) => {
    e.preventDefault();
    focus();
  });
  window.addEventListener("focus", focus);
  focus();
}

export function placeTextarea(ta: HTMLTextAreaElement, x: number, y: number) {
  ta.style.left = `${x}px`;
  ta.style.top = `${y}px`;
}
