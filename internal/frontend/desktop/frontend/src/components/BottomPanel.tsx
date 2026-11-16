import { useEffect, useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import { invoke, problemsStore } from "../bridge";
import type { UIState } from "../types";
import { baseName } from "./TopBar";

const HEIGHT_KEY = "gotextforge.bottomHeight";

function savedHeight() {
  try {
    const v = Number(localStorage.getItem(HEIGHT_KEY));
    return v >= 120 ? v : 240;
  } catch {
    return 240;
  }
}

export function useBottomVisible(ui: UIState | null) {
  const problems = useSyncExternalStore(problemsStore.subscribe, problemsStore.get);
  const bottom = ui?.Bottom ?? "";
  const first = useRef(true);
  useEffect(() => {
    if (first.current) {
      first.current = false;
      return;
    }
    if (bottom !== "") {
      problemsStore.set(false);
    }
  }, [bottom]);
  return problems || bottom !== "";
}

export function BottomPanel({ ui }: { ui: UIState }) {
  const problemsOpen = useSyncExternalStore(problemsStore.subscribe, problemsStore.get);
  const view = problemsOpen ? "problems" : ui.Bottom;
  const count = ui.Problems?.length ?? 0;
  const [height, setHeight] = useState(savedHeight);

  const startDrag = (e: React.MouseEvent) => {
    const startY = e.clientY;
    const startH = height;
    let last = startH;
    const move = (ev: MouseEvent) => {
      last = Math.min(Math.max(startH + startY - ev.clientY, 120), window.innerHeight - 200);
      setHeight(last);
    };
    const up = () => {
      window.removeEventListener("mousemove", move);
      window.removeEventListener("mouseup", up);
      try {
        localStorage.setItem(HEIGHT_KEY, String(last));
      } catch {}
    };
    window.addEventListener("mousemove", move);
    window.addEventListener("mouseup", up);
  };

  const showTab = (tab: "terminal" | "problems" | "docker") => {
    if (tab === "problems") {
      problemsStore.set(true);
      return;
    }
    problemsStore.set(false);
    if (ui.Bottom !== tab) {
      invoke("TogglePanel", tab);
    }
  };

  const close = () => {
    if (view === "problems") {
      problemsStore.set(false);
    } else if (view === "terminal") {
      invoke("TerminalAction", "close");
    } else if (view === "docker") {
      invoke("DockerKey", "q");
    }
  };

  return (
    <section className="bottom" style={{ height }}>
      <div className="drag" onMouseDown={startDrag} />
      <div className="btabs">
        <span className={view === "terminal" ? "active" : ""} onClick={() => showTab("terminal")}>
          Terminal
        </span>
        <span className={view === "problems" ? "active" : ""} onClick={() => showTab("problems")}>
          Problemas {count > 0 && <b className="count">{count}</b>}
        </span>
        <span className={view === "docker" ? "active" : ""} onClick={() => showTab("docker")}>
          Docker
        </span>
        <span className="btabs-push" />
        <span className="icon-btn" title="Fechar painel" onClick={close}>
          ×
        </span>
      </div>
      {view === "terminal" && <TerminalView ui={ui} />}
      {view === "problems" && <ProblemsView ui={ui} />}
      {view === "docker" && <DockerView ui={ui} />}
    </section>
  );
}

const STATE_LABEL: Record<string, string> = {
  running: "rodando…",
  ok: "✓ concluído",
  failed: "✗ falhou",
  canceled: "■ cancelado",
};

function TerminalView({ ui }: { ui: UIState }) {
  const t = ui.Terminal;
  const lines = t.Lines ?? [];
  const body = useRef<HTMLDivElement>(null);
  const focused = ui.Focus === "terminal";

  useLayoutEffect(() => {
    const el = body.current;
    if (!el) return;
    if (t.Follow) {
      el.scrollTop = el.scrollHeight;
    } else {
      el.querySelector(".line.selected")?.scrollIntoView({ block: "nearest" });
    }
  }, [lines.length, t.Selected, t.Follow, t.First]);

  return (
    <div className="pane">
      <div className="pane-head">
        {t.Command ? (
          <>
            <span className="prompt">➜ {baseName(t.Dir)}</span>
            <span className="cmd">{t.Command}</span>
            {t.State && (
              <span className={`state ${t.State}`}>
                {STATE_LABEL[t.State]}
                {t.State === "failed" && t.ExitCode !== 0 && ` (código ${t.ExitCode})`}
                {t.State !== "running" && t.Millis > 0 && ` · ${(t.Millis / 1000).toFixed(1)}s`}
              </span>
            )}
          </>
        ) : (
          <span className="muted">pasta: {t.Dir}</span>
        )}
        <span className="btabs-push" />
        {t.State === "running" && (
          <button onClick={() => invoke("TerminalAction", "cancel")} title="Ctrl+c">
            ■ Cancelar
          </button>
        )}
        {t.Command && (
          <button onClick={() => invoke("TerminalAction", "rerun")} title="r">
            ↻ Repetir
          </button>
        )}
        <button onClick={() => invoke("TerminalAction", "new")} title="i ou :">
          ＋ Comando
        </button>
      </div>
      <div className={focused ? "term focused" : "term"} ref={body}>
        {lines.length === 0 && (
          <div className="muted empty">
            Nenhum comando ainda. Use <kbd>＋ Comando</kbd>, <kbd>Espaço t n</kbd> ou <kbd>:!go test ./...</kbd>
          </div>
        )}
        {lines.map((l, i) => {
          const idx = t.First + i;
          const kind = t.Kinds[i] ?? " ";
          const cls = ["line", `k-${kind === " " ? "p" : kind}`];
          if (idx === t.Selected && (focused || !t.Follow)) cls.push("selected");
          return (
            <div key={idx} className={cls.join(" ")} onClick={() => invoke("TerminalClick", idx)}>
              {l || " "}
            </div>
          );
        })}
      </div>
    </div>
  );
}

const SEV_ICON: Record<string, string> = { error: "⊗", warning: "△", info: "ⓘ", hint: "·" };

function ProblemsView({ ui }: { ui: UIState }) {
  const problems = ui.Problems ?? [];
  if (problems.length === 0) {
    return <div className="pane muted empty">✓ Nenhum problema nas abas abertas.</div>;
  }
  const groups = new Map<string, typeof problems>();
  for (const p of problems) {
    const key = p.Path || p.Name;
    groups.set(key, [...(groups.get(key) ?? []), p]);
  }
  return (
    <div className="pane list">
      {[...groups.entries()].map(([path, items]) => (
        <div key={path}>
          <div className="group">
            {items[0].Name} <span className="muted">{items.length}</span>
          </div>
          {items.map((p, i) => (
            <div key={i} className="item" onClick={() => invoke("GoTo", p.Path, p.Line, p.Col)} title={p.Path}>
              <span className={`sev ${p.Severity}`}>{SEV_ICON[p.Severity]}</span>
              <span className="msg">{p.Message}</span>
              <span className="muted">
                {p.Line}:{p.Col}
              </span>
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}

const DOCKER_TABS = ["Containers", "Compose", "Imagens"];

const DOCKER_ACTIONS: Record<string, { key: string; label: string }[]> = {
  Containers: [
    { key: "s", label: "▶/■ Start/Stop" },
    { key: "r", label: "↻ Restart" },
    { key: "l", label: "≡ Logs" },
    { key: "d", label: "✕ Remover" },
  ],
  Compose: [
    { key: "b", label: "⚒ Rebuild" },
    { key: "s", label: "▶/■ Start/Stop" },
    { key: "r", label: "↻ Restart" },
    { key: "l", label: "≡ Logs" },
  ],
  Imagens: [{ key: "d", label: "✕ Remover" }],
};

function stateClass(state: string) {
  if (state === "running") return "ok";
  if (state === "exited" || state === "dead") return "bad";
  return "mid";
}

function DockerView({ ui }: { ui: UIState }) {
  const d = ui.Docker;
  const focused = ui.Focus === "docker";
  if (!d.Open) {
    return <div className="pane muted empty">Painel Docker fechado.</div>;
  }

  let rows: { key: string; cells: React.ReactNode[] }[] = [];
  if (d.Tab === "Compose") {
    rows = (d.Services ?? []).map((s) => ({
      key: s.Name,
      cells: [
        <span className={`dot ${stateClass(s.State)}`}>●</span>,
        s.Service,
        <span className={stateClass(s.State)}>{s.State === "exited" ? `exited (${s.ExitCode})` : s.State}</span>,
        s.Ports,
      ],
    }));
  } else if (d.Tab === "Imagens") {
    rows = (d.Images ?? []).map((img) => ({
      key: img.ID + img.Repository + img.Tag,
      cells: ["", img.Tag ? `${img.Repository}:${img.Tag}` : img.Repository, img.Size, img.ID.replace("sha256:", "").slice(0, 12)],
    }));
  } else {
    rows = (d.Containers ?? []).map((c) => ({
      key: c.ID,
      cells: [<span className={`dot ${stateClass(c.State)}`}>●</span>, c.Name, <span className={stateClass(c.State)}>{c.Status}</span>, c.Ports],
    }));
  }

  return (
    <div className="pane">
      <div className="pane-head">
        {DOCKER_TABS.map((tab) => (
          <span key={tab} className={tab === d.Tab ? "chip active" : "chip"} onClick={() => invoke("DockerTab", tab)}>
            {tab}
          </span>
        ))}
        <span className="btabs-push" />
        {rows.length > 0 &&
          DOCKER_ACTIONS[d.Tab]?.map((a) => (
            <button key={a.key} onClick={() => invoke("DockerKey", a.key)} title={`tecla ${a.key}`}>
              {a.label}
            </button>
          ))}
      </div>
      {d.Error ? (
        <div className="muted empty">{d.Error}</div>
      ) : rows.length === 0 ? (
        <div className="muted empty">Nada aqui.</div>
      ) : (
        <div className="list">
          <table className={focused ? "dtable focused" : "dtable"}>
            <tbody>
              {rows.map((r, i) => (
                <tr key={r.key} className={i === d.Selected ? "selected" : ""} onClick={() => invoke("DockerClick", i)}>
                  {r.cells.map((c, j) => (
                    <td key={j}>{c}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
