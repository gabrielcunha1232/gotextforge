import { Call, Events } from "@wailsio/runtime";
import { Frame, Renderer, measureCell } from "./render";
import { KeyInput, attachKeyboard, placeTextarea } from "./keyboard";

const SERVICE = "github.com/gabrielcunha1232/gotextforge/internal/frontend/desktop.EditorService";

const canvas = document.getElementById("screen") as HTMLCanvasElement;
const ta = document.getElementById("keys") as HTMLTextAreaElement;
const ctx = canvas.getContext("2d", { alpha: false })!;
const renderer = new Renderer(canvas, ctx, measureCell(ctx));

let pendingKeys: KeyInput[] = [];
let pendingResize = false;
let busy = false;
let lastSeq = 0;

function grid() {
  return {
    cols: Math.max(1, Math.floor(canvas.clientWidth / renderer.cell.w)),
    rows: Math.max(2, Math.floor(canvas.clientHeight / renderer.cell.h)),
  };
}

function show(frame: Frame) {
  if (frame.seq <= lastSeq) {
    return;
  }
  lastSeq = frame.seq;
  renderer.draw(frame);
  const c = frame.cursor;
  placeTextarea(ta, c.col * renderer.cell.w, c.row * renderer.cell.h);
}

async function pump() {
  if (busy) {
    return;
  }
  busy = true;
  try {
    while (pendingResize || pendingKeys.length > 0) {
      let frame: Frame;
      if (pendingResize) {
        pendingResize = false;
        const { cols, rows } = grid();
        frame = await Call.ByName(`${SERVICE}.Resize`, cols, rows);
      } else {
        const keys = pendingKeys;
        pendingKeys = [];
        frame = await Call.ByName(`${SERVICE}.Keys`, keys);
      }
      show(frame);
    }
  } finally {
    busy = false;
  }
}

function resize() {
  pendingResize = true;
  pump();
}

attachKeyboard(ta, (k) => {
  pendingKeys.push(k);
  pump();
});

Events.On("frame", (ev) => show(ev.data as Frame));

window.addEventListener("resize", resize);
document.fonts.ready.then(async () => {
  renderer.cell = measureCell(ctx);
  resize();
  if (await Call.ByName(`${SERVICE}.BenchMode`)) {
    const { runBench } = await import("./bench");
    await new Promise((r) => setTimeout(r, 4000));
    const { cols, rows } = grid();
    const report = await runBench((keys) => Call.ByName(`${SERVICE}.Keys`, keys), show, cols, rows);
    await Call.ByName(`${SERVICE}.Report`, report);
  }
});
