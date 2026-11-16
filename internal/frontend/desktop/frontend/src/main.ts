const BG = "#0b0e13";
const FG = "#d6e2ec";
const FONT_SIZE = 15;
const FONT = `${FONT_SIZE}px "Cascadia Mono", Consolas, monospace`;

const canvas = document.getElementById("screen") as HTMLCanvasElement;
const ctx = canvas.getContext("2d", { alpha: false })!;

function resize() {
  const dpr = window.devicePixelRatio || 1;
  const w = canvas.clientWidth;
  const h = canvas.clientHeight;
  canvas.width = Math.round(w * dpr);
  canvas.height = Math.round(h * dpr);
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  draw(w, h, dpr);
}

function draw(w: number, h: number, dpr: number) {
  ctx.fillStyle = BG;
  ctx.fillRect(0, 0, w, h);

  ctx.font = FONT;
  ctx.textBaseline = "top";
  ctx.fillStyle = FG;

  const cellW = ctx.measureText("M").width;
  const cellH = Math.ceil(FONT_SIZE * 1.4);
  const cols = Math.floor(w / cellW);
  const rows = Math.floor(h / cellH);

  ctx.fillText("gotextforge — desktop (MS7, passo 1)", cellW * 2, cellH * 1);
  ctx.fillText(`janela ${w}x${h} px · dpr ${dpr}`, cellW * 2, cellH * 3);
  ctx.fillText(`célula ${cellW.toFixed(2)}x${cellH} px · grade ${cols} colunas x ${rows} linhas`, cellW * 2, cellH * 4);
}

window.addEventListener("resize", resize);
document.fonts.ready.then(resize);
resize();
