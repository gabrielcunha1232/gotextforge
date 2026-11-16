import type { UIState } from "../types";

export function BottomPanel({ ui }: { ui: UIState }) {
  const problems = ui.Problems?.length ?? 0;
  return (
    <section className="bottom">
      <div className="btabs">
        <span className={ui.Bottom === "terminal" ? "active" : ""}>Terminal</span>
        <span>
          Problemas {problems > 0 && <b className="count">({problems})</b>}
        </span>
        <span className={ui.Bottom === "docker" ? "active" : ""}>Docker</span>
      </div>
      {ui.Bottom === "terminal" ? <TerminalBody ui={ui} /> : <DockerBody ui={ui} />}
    </section>
  );
}

function TerminalBody({ ui }: { ui: UIState }) {
  const t = ui.Terminal;
  const lines = t.Lines ?? [];
  return (
    <div className="terminal">
      {t.Command && (
        <div>
          <span className="prompt">➜　{t.Dir}</span>　{t.Command}
        </div>
      )}
      {lines.slice(-200).map((l, i) => (
        <div key={i}>{l || " "}</div>
      ))}
    </div>
  );
}

function DockerBody({ ui }: { ui: UIState }) {
  const d = ui.Docker;
  if (d.Error) {
    return <div className="terminal muted">{d.Error}</div>;
  }
  return (
    <div className="terminal">
      {(d.Containers ?? []).map((c) => (
        <div key={c.ID}>
          {c.State === "running" ? "●" : "○"} {c.Name}　<span className="muted">{c.Status}</span>
        </div>
      ))}
    </div>
  );
}
