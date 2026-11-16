import { invoke } from "../bridge";
import type { UIState } from "../types";

export function TopBar({ ui }: { ui: UIState | null }) {
  const project = ui?.Explorer.Name || baseName(ui?.Root ?? "");
  return (
    <header className="top">
      <div className="brand">
        <div className="mark">G</div>
        gotextforge
        {project && <span className="project">{project}</span>}
      </div>
      <div className="search" onClick={() => invoke("TogglePanel", "finder")}>
        <span>⌕　Buscar arquivo…</span>
        <span>
          <kbd>Ctrl K</kbd> <kbd>Espaço Espaço</kbd>
        </span>
      </div>
      <div className="top-spacer" />
    </header>
  );
}

export function baseName(path: string) {
  const parts = path.split(/[\\/]/).filter(Boolean);
  return parts[parts.length - 1] ?? "";
}
