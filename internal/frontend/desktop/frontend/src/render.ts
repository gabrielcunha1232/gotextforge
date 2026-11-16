export const BG = "#0b0e13";
export const FG = "#d6e2ec";
const CURSOR = "#d6e2ec";
const FONT_SIZE = 15;
const FONT_FAMILY = `"Cascadia Mono", Consolas, monospace`;

export type Span = {
  t: string;
  fg?: string;
  bg?: string;
  sp?: string;
  b?: boolean;
  i?: boolean;
  u?: boolean;
  c?: boolean;
};

export type Frame = {
  seq: number;
  lines: Span[][];
  status: Span[];
  cursor: { row: number; col: number; bar?: boolean; hide?: boolean };
};

export type Cell = { w: number; h: number; ascent: number };

function font(s?: Span) {
  return `${s?.i ? "italic " : ""}${s?.b ? "bold " : ""}${FONT_SIZE}px ${FONT_FAMILY}`;
}

export function measureCell(ctx: CanvasRenderingContext2D): Cell {
  ctx.font = font();
  const m = ctx.measureText("M");
  const ascent = Math.ceil(m.fontBoundingBoxAscent);
  const descent = Math.ceil(m.fontBoundingBoxDescent);
  return { w: m.width, h: ascent + descent, ascent };
}

export class Renderer {
  private dpr = 1;

  constructor(
    private canvas: HTMLCanvasElement,
    private ctx: CanvasRenderingContext2D,
    public cell: Cell,
  ) {}

  private snap(v: number) {
    return Math.round(v * this.dpr) / this.dpr;
  }

  private x(col: number) {
    return this.snap(col * this.cell.w);
  }

  private fit() {
    this.dpr = window.devicePixelRatio || 1;
    const w = Math.round(this.canvas.clientWidth * this.dpr);
    const h = Math.round(this.canvas.clientHeight * this.dpr);
    if (this.canvas.width !== w || this.canvas.height !== h) {
      this.canvas.width = w;
      this.canvas.height = h;
    }
    this.ctx.setTransform(this.dpr, 0, 0, this.dpr, 0, 0);
  }

  draw(frame: Frame) {
    this.fit();
    const ctx = this.ctx;
    ctx.fillStyle = BG;
    ctx.fillRect(0, 0, this.canvas.clientWidth, this.canvas.clientHeight);
    ctx.textBaseline = "alphabetic";

    const rows = [...frame.lines, frame.status];
    rows.forEach((spans, row) => this.drawRow(spans, row));

    const c = frame.cursor;
    if (!c.hide && c.row < rows.length) {
      this.drawCursor(rows[c.row], c.row, c.col, !!c.bar);
    }
  }

  private drawRow(spans: Span[], row: number) {
    const ctx = this.ctx;
    const y = row * this.cell.h;

    let col = 0;
    for (const s of spans) {
      const n = [...s.t].length;
      if (s.bg) {
        ctx.fillStyle = s.bg;
        ctx.fillRect(this.x(col), y, this.x(col + n) - this.x(col), this.cell.h);
      }
      col += n;
    }

    col = 0;
    for (const s of spans) {
      const n = [...s.t].length;
      this.drawText(s, s.fg || FG, row, col);
      if (s.u || s.c) {
        this.drawUnderline(s, row, col, n);
      }
      col += n;
    }
  }

  private drawText(s: Span, color: string, row: number, col: number) {
    const ctx = this.ctx;
    const baseline = row * this.cell.h + this.cell.ascent;
    ctx.font = font(s);
    ctx.fillStyle = color;
    ctx.strokeStyle = color;

    let run = "";
    let runCol = col;
    const flush = () => {
      if (run.trim() !== "") {
        ctx.fillText(run, this.x(runCol), baseline);
      }
      run = "";
    };

    for (const ch of s.t) {
      if (this.drawBox(ch, row, col)) {
        flush();
        runCol = col + 1;
      } else {
        run += ch;
      }
      col++;
    }
    flush();
  }

  private drawBox(ch: string, row: number, col: number): boolean {
    const code = ch.codePointAt(0)!;
    const ctx = this.ctx;
    const x0 = this.x(col);
    const x1 = this.x(col + 1);
    const y0 = row * this.cell.h;
    const y1 = y0 + this.cell.h;
    const w = x1 - x0;
    const h = this.cell.h;

    if (code === 0x2580) {
      ctx.fillRect(x0, y0, w, this.snap(h / 2));
      return true;
    }
    if (code === 0x2584) {
      const mid = this.snap(y0 + h / 2);
      ctx.fillRect(x0, mid, w, y1 - mid);
      return true;
    }
    if (code >= 0x2588 && code <= 0x258f) {
      const eighths = 8 - (code - 0x2588);
      ctx.fillRect(x0, y0, this.snap((w * eighths) / 8), h);
      return true;
    }
    if (code === 0x2590) {
      const mid = this.snap(x0 + w / 2);
      ctx.fillRect(mid, y0, x1 - mid, h);
      return true;
    }

    const lines: Record<number, string> = {
      0x2500: "lr",
      0x2502: "tb",
      0x251c: "tbr",
      0x2524: "tbl",
      0x256d: "(rb",
      0x256e: "(lb",
      0x2570: "(rt",
      0x256f: "(lt",
    };
    const spec = lines[code];
    if (!spec) {
      return false;
    }

    const lw = Math.max(1, Math.round(this.dpr)) / this.dpr;
    const cx = this.snap(x0 + w / 2) + lw / 2;
    const cy = this.snap(y0 + h / 2) + lw / 2;
    const end = { l: [x0, cy], r: [x1, cy], t: [cx, y0], b: [cx, y1] } as const;

    ctx.lineWidth = lw;
    ctx.beginPath();
    if (spec[0] === "(") {
      const [ax, ay] = end[spec[1] as "l" | "r"];
      const [bx, by] = end[spec[2] as "t" | "b"];
      ctx.moveTo(ax, ay);
      ctx.quadraticCurveTo(cx, cy, bx, by);
    } else {
      for (const d of spec) {
        const [ex, ey] = end[d as "l" | "r" | "t" | "b"];
        ctx.moveTo(cx, cy);
        ctx.lineTo(ex, ey);
      }
    }
    ctx.stroke();
    return true;
  }

  private drawUnderline(s: Span, row: number, col: number, n: number) {
    const ctx = this.ctx;
    const x0 = this.x(col);
    const x1 = this.x(col + n);
    const y = row * this.cell.h + this.cell.ascent + 2.5;
    ctx.strokeStyle = s.sp || s.fg || FG;
    ctx.lineWidth = 1;
    ctx.beginPath();
    if (s.c) {
      const amp = 1.5;
      const half = 2;
      ctx.moveTo(x0, y);
      let up = true;
      for (let x = x0; x < x1; x += half) {
        const nx = Math.min(x + half, x1);
        ctx.quadraticCurveTo((x + nx) / 2, up ? y - amp * 2 : y + amp * 2, nx, y);
        up = !up;
      }
    } else {
      ctx.moveTo(x0, y);
      ctx.lineTo(x1, y);
    }
    ctx.stroke();
  }

  private drawCursor(spans: Span[], row: number, col: number, bar: boolean) {
    const ctx = this.ctx;
    const x0 = this.x(col);
    const y0 = row * this.cell.h;
    ctx.fillStyle = CURSOR;

    if (bar) {
      ctx.fillRect(x0, y0, Math.max(2, Math.round(this.dpr)) / this.dpr, this.cell.h);
      return;
    }

    ctx.fillRect(x0, y0, this.x(col + 1) - x0, this.cell.h);
    const under = spanAt(spans, col);
    if (under) {
      this.drawText(under.span, BG, row, col);
    }
  }
}

function spanAt(spans: Span[], col: number): { span: Span } | null {
  let c = 0;
  for (const s of spans) {
    const chars = [...s.t];
    if (col < c + chars.length) {
      return { span: { ...s, t: chars[col - c] } };
    }
    c += chars.length;
  }
  return null;
}
