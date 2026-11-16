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
      <div className="search">
        <span>⌕　Buscar arquivo…</span>
        <kbd>Espaço Espaço</kbd>
      </div>
      <div className="top-spacer" />
    </header>
  );
}

export function baseName(path: string) {
  const parts = path.split(/[\\/]/).filter(Boolean);
  return parts[parts.length - 1] ?? "";
}
