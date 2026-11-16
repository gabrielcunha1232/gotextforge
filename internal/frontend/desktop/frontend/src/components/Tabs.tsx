import type { UIState } from "../types";

export function Tabs({ ui }: { ui: UIState | null }) {
  const tabs = ui?.Tabs ?? [];
  return (
    <>
      <div className="tabs">
        {tabs.map((t, i) => (
          <div key={t.Path || i} className={t.Active ? "tab active" : "tab"} title={t.Path}>
            {t.Name}
            {t.Modified && <span className="dot">●</span>}
          </div>
        ))}
      </div>
      <div className="crumb">{crumbs(ui)}</div>
    </>
  );
}

function crumbs(ui: UIState | null) {
  const path = ui?.Status.Path;
  if (!ui || !path) {
    return null;
  }
  let rel = path;
  const root = ui.Root.replace(/[\\/]+$/, "");
  if (root && rel.toLowerCase().startsWith(root.toLowerCase())) {
    rel = rel.slice(root.length);
  }
  const parts = rel.split(/[\\/]/).filter(Boolean);
  return parts.map((p, i) => (
    <span key={i} className={i === parts.length - 1 ? "last" : ""}>
      {i > 0 && <span className="sep">›</span>}
      {p}
    </span>
  ));
}
