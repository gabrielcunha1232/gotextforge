import { Call, Events } from "@wailsio/runtime";
import { Frame, Renderer, measureCell } from "./render";
import { KeyInput, attachKeyboard, placeTextarea } from "./keyboard";
import type { Pos, UIState } from "./types";

const SERVICE = "github.com/gabrielcunha1232/gotextforge/internal/frontend/desktop.EditorService";

export function call<T>(method: string, ...args: unknown[]): Promise<T> {
  return Call.ByName(`${SERVICE}.${method}`, ...args) as Promise<T>;
}

type Listener = () => void;

class Store<T> {
  private listeners = new Set<Listener>();
  constructor(public value: T) {}
  set(v: T) {
    this.value = v;
    this.listeners.forEach((l) => l());
  }
  subscribe = (l: Listener) => {
    this.listeners.add(l);
    return () => this.listeners.delete(l);
  };
  get = () => this.value;
}

export const uiStore = new Store<UIState | null>(null);
export const posStore = new Store<Pos>({ line: 1, col: 1, progress: "" });
export const problemsStore = new Store<boolean>(false);

let renderer: Renderer | null = null;
let canvas: HTMLCanvasElement | null = null;
let textarea: HTMLTextAreaElement | null = null;

type Op = { method: string; args: unknown[] };

let ops: Op[] = [];
let pendingResize: "start" | "resize" | null = null;
let busy = false;
let lastSeq = 0;
let lastUISeq = 0;

function grid() {
  const c = canvas!;
  const cell = renderer!.cell;
  return {
    cols: Math.max(1, Math.floor(c.clientWidth / cell.w)),
    rows: Math.max(2, Math.floor(c.clientHeight / cell.h)),
  };
}

function show(frame: Frame) {
  if (frame.ui && frame.seq > lastUISeq) {
    lastUISeq = frame.seq;
    uiStore.set(frame.ui);
  }
  if (frame.seq <= lastSeq || !renderer) {
    return;
  }
  lastSeq = frame.seq;
  renderer.draw(frame);
  posStore.set(frame.pos);
  const c = frame.cursor;
  placeTextarea(textarea!, c.col * renderer.cell.w, c.row * renderer.cell.h);
}

async function pump() {
  if (busy || !renderer) {
    return;
  }
  busy = true;
  try {
    while (pendingResize || ops.length > 0) {
      let frame: Frame;
      if (pendingResize) {
        const method = pendingResize === "start" ? "Start" : "Resize";
        pendingResize = null;
        const { cols, rows } = grid();
        frame = await call<Frame>(method, cols, rows);
      } else {
        const op = ops.shift()!;
        frame = await call<Frame>(op.method, ...op.args);
      }
      show(frame);
    }
  } finally {
    busy = false;
  }
}

export function sendKey(k: KeyInput) {
  const last = ops[ops.length - 1];
  if (last && last.method === "Keys") {
    (last.args[0] as KeyInput[]).push(k);
  } else {
    ops.push({ method: "Keys", args: [[k]] });
  }
  pump();
}

export function invoke(method: string, ...args: unknown[]) {
  ops.push({ method, args });
  pump();
}

export async function loadTheme() {
  const theme = await call<Record<string, string>>("Theme");
  for (const [name, color] of Object.entries(theme)) {
    document.documentElement.style.setProperty(`--c-${name}`, color);
  }
}

export function attachEditor(c: HTMLCanvasElement, ta: HTMLTextAreaElement) {
  canvas = c;
  textarea = ta;
  const ctx = c.getContext("2d", { alpha: false })!;
  renderer = new Renderer(c, ctx, measureCell(ctx));

  attachKeyboard(ta, sendKey);
  Events.On("frame", (ev) => show(ev.data as Frame));

  const observer = new ResizeObserver(() => {
    if (pendingResize !== "start") {
      pendingResize = "resize";
    }
    pump();
  });

  document.fonts.ready.then(async () => {
    renderer!.cell = measureCell(ctx);
    pendingResize = "start";
    observer.observe(c);
    await pump();
    if (await call<boolean>("BenchMode")) {
      const { runBench } = await import("./bench");
      await new Promise((r) => setTimeout(r, 4000));
      const { cols, rows } = grid();
      const report = await runBench((keys) => call<Frame>("Keys", keys), show, cols, rows);
      await call("Report", report);
    }
  });
}
