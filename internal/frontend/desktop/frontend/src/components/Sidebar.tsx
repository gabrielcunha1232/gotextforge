import type { UIState } from "../types";

export function Sidebar({ ui }: { ui: UIState | null }) {
  const x = ui?.Explorer;
  const rows = x?.Rows ?? [];
  const focused = ui?.Focus === "explorer";

  return (
    <aside className="side">
      <div className="cap">Explorador</div>
      {rows.length === 0 ? (
        <div className="hint">
          <kbd>Espaço e</kbd> abre o explorador
        </div>
      ) : (
        <>
          <div className="proj">⌄　{x!.Name}</div>
          <div className="tree">
            {rows.map((r, i) => {
              const cls = ["row"];
              if (r.Dir) cls.push("dir");
              if (r.Current) cls.push("current");
              if (i === x!.Selected) cls.push(focused ? "sel-focus" : "sel");
              return (
                <div key={r.Path} className={cls.join(" ")} style={{ paddingLeft: 8 + r.Depth * 14 }}>
                  <span className="twisty">{r.Dir ? (r.Expanded ? "▾" : "▸") : ""}</span>
                  {r.Name}
                  {r.Dir && r.Expanded && r.Empty && <span className="muted"> (vazia)</span>}
                </div>
              );
            })}
          </div>
        </>
      )}
    </aside>
  );
}
