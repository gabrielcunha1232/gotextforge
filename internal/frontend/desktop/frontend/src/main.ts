import { Call } from "@wailsio/runtime";

const SERVICE = "github.com/gabrielcunha1232/gotextforge/internal/frontend/desktop.EditorService";
const BG = "#0b0e13";
const FG = "#d6e2ec";
const FONT_SIZE = 15;
const FONT = `${FONT_SIZE}px "Cascadia Mono", Consolas, monospace`;

type Frame = {
  lines: string[];
  status: string;
  cursor: { row: number; col: number };
};

type Cell = { w: number; h: number; ascent: number };

const canvas = document.getElementById("screen") as HTMLCanvasElement;
const ctx = canvas.getContext("2d", { alpha: false })!;

let cell: Cell = measureCell();
let lastSeq = 0;

function measureCell(): Cell {
  ctx.font = FONT;
  const m = ctx.measureText("M");
  const ascent = Math.ceil(m.fontBoundingBoxAscent);
  const descent = Math.ceil(m.fontBoundingBoxDescent);
  return { w: m.width, h: ascent + descent, ascent };
}

function grid() {
  return {
    cols: Math.max(1, Math.floor(canvas.clientWidth / cell.w)),
    rows: Math.max(2, Math.floor(canvas.clientHeight / cell.h)),
  };
}

function fitCanvas() {
  const dpr = window.devicePixelRatio || 1;
  const w = Math.round(canvas.clientWidth * dpr);
  const h = Math.round(canvas.clientHeight * dpr);
  if (canvas.width !== w || canvas.height !== h) {
    canvas.width = w;
    canvas.height = h;
  }
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
}

function draw(frame: Frame) {
  fitCanvas();
  ctx.fillStyle = BG;
  ctx.fillRect(0, 0, canvas.clientWidth, canvas.clientHeight);

  ctx.font = FONT;
  ctx.textBaseline = "alphabetic";
  ctx.fillStyle = FG;

  frame.lines.forEach((line, row) => {
    ctx.fillText(line, 0, row * cell.h + cell.ascent);
  });
  ctx.fillText(frame.status, 0, frame.lines.length * cell.h + cell.ascent);
}

async function resize() {
  const seq = ++lastSeq;
  const { cols, rows } = grid();
  const frame: Frame = await Call.ByName(`${SERVICE}.Resize`, cols, rows);
  if (seq === lastSeq) {
    draw(frame);
  }
}

window.addEventListener("resize", resize);
document.fonts.ready.then(() => {
  cell = measureCell();
  resize();
});
