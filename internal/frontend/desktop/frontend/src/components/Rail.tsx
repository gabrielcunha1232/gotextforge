import { invoke } from "../bridge";
import type { UIState } from "../types";

const ITEMS = [
  { id: "explorer", icon: "▤", title: "Explorador (Espaço e)" },
  { id: "finder", icon: "⌕", title: "Buscar arquivo (Espaço Espaço)" },
  { id: "terminal", icon: "▷", title: "Terminal (Espaço t t)" },
  { id: "docker", icon: "◳", title: "Docker (Espaço d o)" },
];

export function Rail({ ui }: { ui: UIState | null }) {
  const focus = ui?.Focus ?? "editor";
  return (
    <aside className="rail">
      {ITEMS.map((it) => {
        const open =
          (it.id === "explorer" && ui?.Explorer.Open) || (ui?.Bottom === it.id) || (it.id === "finder" && ui?.Finder);
        const cls = [focus === it.id ? "sel" : "", open ? "open" : ""].join(" ");
        return (
          <span key={it.id} className={cls} title={it.title} onClick={() => invoke("TogglePanel", it.id)}>
            {it.icon}
          </span>
        );
      })}
      <span className="rail-bottom" title="Configurações (em breve)">
        ⚙
      </span>
    </aside>
  );
}
