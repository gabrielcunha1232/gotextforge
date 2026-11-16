import type { Frame } from "./render";
import type { KeyInput } from "./keyboard";

type Send = (keys: KeyInput[]) => Promise<Frame>;
type Draw = (f: Frame) => void;

function stats(name: string, xs: number[]) {
  const s = [...xs].sort((a, b) => a - b);
  const p = (q: number) => s[Math.min(s.length - 1, Math.floor(q * s.length))].toFixed(2);
  return `${name}: n=${s.length} p50=${p(0.5)}ms p95=${p(0.95)}ms p99=${p(0.99)}ms max=${s[s.length - 1].toFixed(2)}ms`;
}

async function measure(send: Send, draw: Draw, keys: KeyInput[]) {
  const total: number[] = [];
  const call: number[] = [];
  const paint: number[] = [];
  let bytes = 0;
  for (const k of keys) {
    const t0 = performance.now();
    const f = await send([k]);
    const t1 = performance.now();
    draw(f);
    const tDraw = performance.now();
    await new Promise(requestAnimationFrame);
    const t2 = performance.now();
    call.push(t1 - t0);
    paint.push(tDraw - t1);
    total.push(t2 - t0);
    bytes = Math.max(bytes, JSON.stringify(f).length);
  }
  return { total, call, paint, bytes };
}

export async function runBench(send: Send, draw: Draw, cols: number, rows: number): Promise<string> {
  const out = [`grade ${cols}x${rows}`];

  const scroll = Array.from({ length: 300 }, () => ({ key: "j" }));
  let r = await measure(send, draw, scroll);
  out.push("-- rolar (j x300)", stats("tecla→tela", r.total), stats("  chamada Go", r.call), stats("  desenho (sem vsync)", r.paint), `  maior quadro ${r.bytes} B`);

  await send(["2", "5", "0", "0", "G", "A"].map((key) => ({ key })));
  const typing: KeyInput[] = [];
  for (let i = 0; i < 300; i++) {
    typing.push({ key: "x" });
    typing.push({ key: "Backspace" });
  }
  r = await measure(send, draw, typing);
  out.push("-- digitar no meio (x/BS x600)", stats("tecla→tela", r.total), stats("  chamada Go", r.call), stats("  desenho (sem vsync)", r.paint));

  const jump: KeyInput[] = [];
  for (let i = 0; i < 50; i++) {
    jump.push({ key: "Escape" }, { key: "G" }, { key: "g" }, { key: "g" });
  }
  r = await measure(send, draw, jump);
  out.push("-- saltar (G/gg x50)", stats("tecla→tela", r.total));

  await send([{ key: "Escape" }, { key: "u" }]);
  return out.join("\n");
}
