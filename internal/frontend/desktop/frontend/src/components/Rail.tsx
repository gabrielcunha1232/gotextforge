import type { UIState } from "../types";

const ITEMS = [
  { id: "explorer", icon: "▤", title: "Explorador (Espaço e)" },
  { id: "finder", icon: "⌕", title: "Buscar arquivo (Espaço Espaço)" },
  { id: "terminal", icon: "▷", title: "Terminal (Espaço t t)" },
  { id: "docker", icon: "◳", title: "Docker (Espaço d o)" },
];

export function Rail({ focus, bottom }: { focus: UIState["Focus"]; bottom: UIState["Bottom"] }) {
  return (
    <aside className="rail">
      {ITEMS.map((it) => {
        const active = focus === it.id || bottom === it.id;
        return (
          <span key={it.id} className={active ? "sel" : ""} title={it.title}>
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
