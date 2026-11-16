import { Call } from "@wailsio/runtime";
import { Frame, Renderer, measureCell } from "./render";

const SERVICE = "github.com/gabrielcunha1232/gotextforge/internal/frontend/desktop.EditorService";

const canvas = document.getElementById("screen") as HTMLCanvasElement;
const ctx = canvas.getContext("2d", { alpha: false })!;
const renderer = new Renderer(canvas, ctx, measureCell(ctx));

let lastSeq = 0;

function grid() {
  return {
    cols: Math.max(1, Math.floor(canvas.clientWidth / renderer.cell.w)),
    rows: Math.max(2, Math.floor(canvas.clientHeight / renderer.cell.h)),
  };
}

async function resize() {
  const seq = ++lastSeq;
  const { cols, rows } = grid();
  const frame: Frame = await Call.ByName(`${SERVICE}.Resize`, cols, rows);
  if (seq === lastSeq) {
    renderer.draw(frame);
  }
}

window.addEventListener("resize", resize);
document.fonts.ready.then(() => {
  renderer.cell = measureCell(ctx);
  resize();
});
