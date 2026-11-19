import { invoke } from "../bridge";
import type { UIState } from "../types";

export function TopBar({ ui }: { ui: UIState | null }) {
  const project = ui?.Explorer.Name || baseName(ui?.Root ?? "");
  const retro = ui?.Theme === "retro";
  return (
    <>
      <header className="top">
        <div className="brand">
          <div className="mark">G</div>
          gotextforge
          {project && (
            <span className="project click" title={`${ui?.Root}
Clique para trocar de pasta (Ctrl Shift O)`} onClick={() => invoke("OpenFolder")}>
              {project} ▾
            </span>
          )}
        </div>
        <div className="search" onClick={() => invoke("TogglePanel", "finder")}>
          <span>⌕　Buscar arquivo…</span>
          <span>
            <kbd>Ctrl K</kbd> <kbd>Espaço Espaço</kbd>
          </span>
        </div>
        <div className="top-spacer">
          <button
            className="theme-btn"
            title={`Trocar para o tema ${retro ? "Go" : "Retrô"} (:colorscheme ${retro ? "go" : "retro"})`}
            onClick={() => invoke("SetTheme", retro ? "go" : "retro")}
          >
            ◐ Tema
          </button>
        </div>
      </header>
      <div className="stripes" />
    </>
  );
}

export function baseName(path: string) {
  const parts = path.split(/[\\/]/).filter(Boolean);
  return parts[parts.length - 1] ?? "";
}
