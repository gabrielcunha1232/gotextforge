import { Call } from "@wailsio/runtime";
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

function grid() {
  return {
    cols: Math.max(1, Math.floor(canvas.clientWidth / renderer.cell.w)),
    rows: Math.max(2, Math.floor(canvas.clientHeight / renderer.cell.h)),
  };
}

function show(frame: Frame) {
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

window.addEventListener("resize", resize);
document.fonts.ready.then(() => {
  renderer.cell = measureCell(ctx);
  resize();
});
